package migration

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type Migrator struct {
	logger         *zap.Logger
	migrationsPath string
}

func New(logger *zap.Logger, migrationsPath string) *Migrator {
	return &Migrator{
		logger:         logger,
		migrationsPath: migrationsPath,
	}
}

func (m *Migrator) Up(databaseURL string) error {
	m.logger.Info("Starting database migrations up")

	absPath, err := filepath.Abs(m.migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	migrationURL := fmt.Sprintf("file://%s", absPath)
	migrator, err := migrate.New(migrationURL, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	version, dirty, err := migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		m.logger.Info("No migrations applied yet")
	} else {
		m.logger.Info("Current migration version", zap.Uint("version", version), zap.Bool("dirty", dirty))
		if dirty {
			m.logger.Warn("Database is in dirty state, forcing cleanup")
			if err := migrator.Force(int(version)); err != nil {
				return fmt.Errorf("failed to force clean dirty state: %w", err)
			}
		}
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	finalVersion, _, err := migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get final migration version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		m.logger.Info("No migrations to apply")
	} else {
		m.logger.Info("Migrations completed successfully", zap.Uint("final_version", finalVersion))
	}

	return nil
}

func (m *Migrator) Down(databaseURL string) error {
	m.logger.Info("Starting database migrations down")

	absPath, err := filepath.Abs(m.migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	migrationURL := fmt.Sprintf("file://%s", absPath)
	migrator, err := migrate.New(migrationURL, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	m.logger.Info("Migrations rollback completed successfully")
	return nil
}
