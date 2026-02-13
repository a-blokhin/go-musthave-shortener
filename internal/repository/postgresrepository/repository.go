package postgresrepository

import (
	"context"
	"errors"
	"fmt"
	"go-musthave-shortener/internal/model"
	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type PostgresRepo struct {
	pool        *pgxpool.Pool
	logger      *zap.Logger
	aliasLength int
}

func New(pool *pgxpool.Pool, logger *zap.Logger) *PostgresRepo {
	return &PostgresRepo{
		pool:        pool,
		logger:      logger,
		aliasLength: 8,
	}
}

func (r *PostgresRepo) Add(ctx context.Context, url string) (string, error) {
	const maxAttempts = 10
	for range maxAttempts {
		alias := generateAlias(r.aliasLength)

		_, err := r.pool.Exec(ctx,
			"INSERT INTO urls (id, short_url, original_url) VALUES ($1, $2, $3)",
			uuid.New(), alias, url)

		if err == nil {
			return alias, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == "idx_urls_original_url_unique" {
				var existingShortURL string
				queryErr := r.pool.QueryRow(ctx,
					"SELECT short_url FROM urls WHERE original_url = $1", url,
				).Scan(&existingShortURL)
				if queryErr != nil {
					r.logger.Error("Failed to query existing short URL after duplicate", zap.Error(queryErr))
					return "", fmt.Errorf("failed to query existing short URL: %w", queryErr)
				}

				return "", &model.DuplicateURLError{ExistingShortURL: existingShortURL}
			}
			continue
		}

		r.logger.Error("Failed to insert URL", zap.Error(err))
		return "", fmt.Errorf("failed to insert URL: %w", err)
	}

	return "", errors.New("failed to generate unique alias after maximum attempts")
}

func (r *PostgresRepo) AddBatch(ctx context.Context, urls []string) ([]string, error) {
	if len(urls) == 0 {
		return []string{}, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result := make([]string, len(urls))

	existing := make(map[string]string, len(urls))
	rows, err := tx.Query(ctx,
		`SELECT original_url, short_url FROM urls WHERE original_url = ANY($1)`,
		urls,
	)
	if err != nil {
		r.logger.Error("Failed to select existing URLs", zap.Error(err))
		return nil, fmt.Errorf("failed to select existing URLs: %w", err)
	}
	for rows.Next() {
		var orig, short string
		if err := rows.Scan(&orig, &short); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan existing URLs: %w", err)
		}
		existing[orig] = short
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed during existing URLs iteration: %w", err)
	}
	rows.Close()

	type ins struct {
		id    uuid.UUID
		alias string
		url   string
	}
	inserts := make([]ins, 0, len(urls))

	generated := make(map[string]string, len(urls))

	for i, u := range urls {
		if short, ok := existing[u]; ok {
			result[i] = short
			continue
		}
		if short, ok := generated[u]; ok {
			result[i] = short
			continue
		}

		alias := generateAlias(r.aliasLength)
		result[i] = alias
		generated[u] = alias
		inserts = append(inserts, ins{id: uuid.New(), alias: alias, url: u})
	}

	var b pgx.Batch
	for _, it := range inserts {
		b.Queue(
			"INSERT INTO urls (id, short_url, original_url) VALUES ($1, $2, $3)",
			it.id, it.alias, it.url,
		)
	}

	br := tx.SendBatch(ctx, &b)
	for range inserts {
		_, err := br.Exec()
		if err != nil {
			_ = br.Close()
			if isUniqueConstraintError(err) {
				return nil, errors.New("collision: batch insert failed due to unique constraint")
			}
			r.logger.Error("Failed to insert URL in batch", zap.Error(err))
			return nil, fmt.Errorf("failed to insert URL in batch: %w", err)
		}
	}
	if err := br.Close(); err != nil {
		return nil, fmt.Errorf("failed to close batch: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

func (r *PostgresRepo) Get(ctx context.Context, alias string) (string, error) {
	var originalURL string
	err := r.pool.QueryRow(ctx, "SELECT original_url FROM urls WHERE short_url = $1", alias).Scan(&originalURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("can't find requested alias %s", alias)
		}
		r.logger.Error("Failed to get URL by alias", zap.String("alias", alias), zap.Error(err))
		return "", fmt.Errorf("failed to get URL by alias: %w", err)
	}

	return originalURL, nil
}

func generateAlias(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}

func isUniqueConstraintError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
