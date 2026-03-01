package postgresrepository

import (
	"context"
	"errors"
	"fmt"
	"go-musthave-shortener/internal/model"
	"go-musthave-shortener/internal/repository"
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

func (r *PostgresRepo) Add(ctx context.Context, url string, userID string) (string, error) {
	const maxAttempts = 10
	for range maxAttempts {
		alias := generateAlias(r.aliasLength)

		_, err := r.pool.Exec(ctx,
			"INSERT INTO urls (id, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)",
			uuid.New(), alias, url, userID)

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

func (r *PostgresRepo) AddBatch(ctx context.Context, urls []string, userID string) ([]string, error) {
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
		`SELECT original_url, short_url FROM urls WHERE original_url = ANY($1) AND (user_id = $2 OR user_id IS NULL)`,
		urls, userID,
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
			"INSERT INTO urls (id, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)",
			it.id, it.alias, it.url, userID,
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
	var isDeleted bool
	err := r.pool.QueryRow(ctx, "SELECT original_url, is_deleted FROM urls WHERE short_url = $1", alias).Scan(&originalURL, &isDeleted)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("can't find requested alias %s", alias)
		}
		r.logger.Error("Failed to get URL by alias", zap.String("alias", alias), zap.Error(err))
		return "", fmt.Errorf("failed to get URL by alias: %w", err)
	}

	if isDeleted {
		return "", &model.DeletedURLError{}
	}

	return originalURL, nil
}

func (r *PostgresRepo) GetByUserID(ctx context.Context, userID string) ([]repository.UserURL, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT short_url, original_url FROM urls WHERE user_id = $1 AND is_deleted = FALSE ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		r.logger.Error("Failed to query URLs by user ID", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to query URLs by user ID: %w", err)
	}
	defer rows.Close()

	var userURLs []repository.UserURL
	for rows.Next() {
		var userURL repository.UserURL
		if err := rows.Scan(&userURL.ShortURL, &userURL.OriginalURL); err != nil {
			r.logger.Error("Failed to scan user URL", zap.Error(err))
			return nil, fmt.Errorf("failed to scan user URL: %w", err)
		}
		userURLs = append(userURLs, userURL)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error during user URLs iteration", zap.Error(err))
		return nil, fmt.Errorf("error during user URLs iteration: %w", err)
	}

	return userURLs, nil
}

func (r *PostgresRepo) BatchDelete(ctx context.Context, shortURLs []string, userID string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	query := "UPDATE urls SET is_deleted = TRUE WHERE short_url = ANY($1) AND user_id = $2 AND is_deleted = FALSE"
	
	_, err := r.pool.Exec(ctx, query, shortURLs, userID)
	if err != nil {
		r.logger.Error("Failed to batch update URLs as deleted", zap.Error(err))
		return fmt.Errorf("failed to batch update URLs as deleted: %w", err)
	}

	return nil
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
