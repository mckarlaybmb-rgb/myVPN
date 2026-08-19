package database

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil { return nil, fmt.Errorf("create database pool: %w", err) }
	if err := pool.Ping(ctx); err != nil { pool.Close(); return nil, fmt.Errorf("ping database: %w", err) }
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool, directory string) error {
	entries, err := os.ReadDir(directory)
	if err != nil { return fmt.Errorf("read migrations: %w", err) }
	files := make([]string, 0, len(entries))
	for _, entry := range entries { if !entry.IsDir() && len(entry.Name()) > 4 && entry.Name()[len(entry.Name())-4:] == ".sql" { files = append(files, entry.Name()) } }
	sort.Strings(files)
	for _, name := range files {
		sql, err := os.ReadFile(directory + "/" + name)
		if err != nil { return fmt.Errorf("read migration %s: %w", name, err) }
		if _, err := pool.Exec(ctx, string(sql)); err != nil { return fmt.Errorf("run migration %s: %w", name, err) }
	}
	return nil
}