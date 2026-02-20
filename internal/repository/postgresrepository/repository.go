package postgresrepository

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
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

func (r *PostgresRepo) Add(url string) (string, error) {
	ctx := context.Background()

	var existingShortURL string
	err := r.pool.QueryRow(ctx, "SELECT short_url FROM urls WHERE original_url = $1", url).Scan(&existingShortURL)
	if err == nil {
		return existingShortURL, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		r.logger.Error("Failed to check existing URL", zap.Error(err))
		return "", fmt.Errorf("failed to check existing URL: %w", err)
	}

	const maxAttempts = 10
	for range maxAttempts {
		alias := generateAlias(r.aliasLength)

		_, err = r.pool.Exec(ctx,
			"INSERT INTO urls (id, short_url, original_url) VALUES ($1, $2, $3)",
			uuid.New(), alias, url)

		if err == nil {
			return alias, nil
		}

		if isUniqueConstraintError(err) {
			continue
		}

		r.logger.Error("Failed to insert URL", zap.Error(err))
		return "", fmt.Errorf("failed to insert URL: %w", err)
	}

	return "", errors.New("failed to generate unique alias after maximum attempts")
}

func (r *PostgresRepo) Get(alias string) (string, error) {
	ctx := context.Background()

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
