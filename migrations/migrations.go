package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"sort"
)

//go:embed *.sql
var Files embed.FS

func RunMigrations(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	entries, err := Files.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		logger.Info("applying migration", slog.String("file", name)) // Добавь логгер в аргументы
		content, err := Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", name, err)
		}
		logger.Info("migration applied successfully", slog.String("file", name))
	}

	return nil
}
