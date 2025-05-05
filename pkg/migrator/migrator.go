package migrator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Migrator interface {
	Up() error
	Down(steps int) error
}

func NewPgx(db *pgx.Conn, path string) Migrator {
	return &pgxMigrator{db: db, path: path}
}

const migrationsTableName string = "schema_migrations"

type pgxMigrator struct {
	db   *pgx.Conn
	path string
}

func (m *pgxMigrator) Up() error {
	if err := m.ensureSchemaMigrationsTable(); err != nil {
		return err
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	files, err := m.getMigrationFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		if _, ok := applied[file]; ok {
			continue
		}

		upSQL, _, err := parseMigrationFile(filepath.Join(m.path, file))
		if err != nil {
			return err
		}

		tx, err := m.db.Begin(context.TODO())
		if err != nil {
			return err
		}

		_, err = tx.Exec(context.TODO(), upSQL)
		if err != nil {
			tx.Rollback(context.TODO())
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}

		_, err = tx.Exec(context.TODO(),
			fmt.Sprintf(`INSERT INTO %s (filename) VALUES ($1)`, migrationsTableName), file)
		if err != nil {
			tx.Rollback(context.TODO())
			return err
		}

		if err := tx.Commit(context.TODO()); err != nil {
			return err
		}

		fmt.Printf("Applied migration: %s\n", file)
	}

	return nil
}

func (m *pgxMigrator) Down(steps int) error {
	if steps <= 0 {
		return errors.New("steps must be > 0")
	}

	var query = fmt.Sprintf(`SELECT filename FROM %s ORDER BY applied_at DESC LIMIT $1`, migrationsTableName)
	rows, err := m.db.Query(context.TODO(), query, steps)
	if err != nil {
		return err
	}
	defer rows.Close()

	var filenames []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return err
		}
		filenames = append(filenames, filename)
	}

	for _, file := range filenames {
		_, downSQL, err := parseMigrationFile(filepath.Join(m.path, file))
		if err != nil {
			return fmt.Errorf("failed to read down migration %s: %w", file, err)
		}

		tx, err := m.db.Begin(context.TODO())
		if err != nil {
			return err
		}

		_, err = tx.Exec(context.TODO(), downSQL)
		if err != nil {
			tx.Rollback(context.TODO())
			return fmt.Errorf("failed to rollback migration %s: %w", file, err)
		}

		_, err = tx.Exec(context.TODO(), fmt.Sprintf(`DELETE FROM %s WHERE filename = $1`, migrationsTableName), file)
		if err != nil {
			tx.Rollback(context.TODO())
			return err
		}

		if err := tx.Commit(context.TODO()); err != nil {
			return err
		}

		fmt.Printf("Rolled back migration: %s\n", file)
	}

	return nil
}

func (m *pgxMigrator) ensureSchemaMigrationsTable() error {
	_, err := m.db.Exec(
		context.TODO(),
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				filename TEXT PRIMARY KEY,
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, migrationsTableName),
	)
	return err
}

func (m *pgxMigrator) getAppliedMigrations() (map[string]struct{}, error) {
	applied := make(map[string]struct{})
	rows, err := m.db.Query(context.TODO(), fmt.Sprintf(`SELECT filename FROM %s`, migrationsTableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = struct{}{}
	}

	return applied, nil
}

func (m *pgxMigrator) getMigrationFiles() ([]string, error) {
	entries, err := os.ReadDir(m.path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func parseMigrationFile(path string) (upSQL string, downSQL string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	sections := strings.Split(string(data), "-- +")
	var up, down string

	for _, section := range sections {
		if strings.HasPrefix(section, "up") {
			up = strings.TrimSpace(strings.TrimPrefix(section, "up"))
		} else if strings.HasPrefix(section, "down") {
			down = strings.TrimSpace(strings.TrimPrefix(section, "down"))
		}
	}

	if up == "" {
		return "", "", fmt.Errorf("no up migration found in %s", path)
	}

	if down == "" {
		return "", "", fmt.Errorf("no down migration found in %s", path)
	}

	return up, down, nil
}
