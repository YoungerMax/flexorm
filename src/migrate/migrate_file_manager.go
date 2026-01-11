package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"flexorm/common/v2"
)

// LocalFileManager manages migration files on the local filesystem
type LocalFileManager struct {
	Config Config
}

// CreateMigration generates a new migration from the current schema
func (m LocalFileManager) CreateMigration(migration Migration) error {
	migrationPath := filepath.Join(m.Config.MigrationsDir, strconv.Itoa(migration.ID))

	// Create migration directory
	if err := os.MkdirAll(migrationPath, 0755); err != nil {
		return fmt.Errorf("failed to create migration directory: %w", err)
	}

	// Create up.sql
	upPath := filepath.Join(migrationPath, "up.sql")
	if err := os.WriteFile(upPath, []byte(migration.UpSQL), 0644); err != nil {
		return fmt.Errorf("failed to create up.sql: %w", err)
	}

	// Create down.sql
	downPath := filepath.Join(migrationPath, "down.sql")
	if err := os.WriteFile(downPath, []byte(migration.DownSQL), 0644); err != nil {
		return fmt.Errorf("failed to create down.sql: %w", err)
	}

	// Create the schema snapshot
	if err := m.SaveSchemaSnapshot(migration.ID, migration.SchemaSnapshot); err != nil {
		return err
	}

	return nil
}

// GetMigration loads a migration from disk
func (m LocalFileManager) GetMigration(id int) (*Migration, error) {
	migrationPath := filepath.Join(m.Config.MigrationsDir, strconv.Itoa(id))

	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migration %d does not exist", id)
	}

	// Read up.sql
	upPath := filepath.Join(migrationPath, "up.sql")
	upSQL, err := os.ReadFile(upPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read up.sql: %w", err)
	}

	// Read down.sql
	downPath := filepath.Join(migrationPath, "down.sql")
	downSQL, err := os.ReadFile(downPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read down.sql: %w", err)
	}

	// Load the schema snapshot
	schemaSnapshot, err := m.LoadSchemaSnapshot(id)
	if err != nil {
		return nil, err
	}

	return &Migration{
		ID:             id,
		UpSQL:          string(upSQL),
		DownSQL:        string(downSQL),
		SchemaSnapshot: *schemaSnapshot,
	}, nil
}

// ListMigrations lists all migrations on disk
func (m LocalFileManager) ListMigrations() ([]Migration, error) {
	entries, err := os.ReadDir(m.Config.MigrationsDir)

	if err != nil {
		if os.IsNotExist(err) {
			return []Migration{}, nil
		}

		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id, err := strconv.Atoi(entry.Name())
		if err != nil {
			return []Migration{}, err
		}

		migration, err := m.GetMigration(id)
		if err != nil {
			return []Migration{}, err
		}

		migrations = append(migrations, *migration)
	}

	// Sort by ID
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

// SaveSchemaSnapshot saves a schema snapshot to disk
func (m LocalFileManager) SaveSchemaSnapshot(id int, schema common.Schema) error {
	migrationPath := filepath.Join(m.Config.MigrationsDir, strconv.Itoa(id))
	snapshotPath := filepath.Join(migrationPath, "schema_snapshot.json")

	data, err := json.MarshalIndent(schema, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	if err := os.WriteFile(snapshotPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write schema snapshot: %w", err)
	}

	return nil
}

// LoadSchemaSnapshot loads a schema snapshot from disk
func (m LocalFileManager) LoadSchemaSnapshot(id int) (*common.Schema, error) {
	migrationPath := filepath.Join(m.Config.MigrationsDir, strconv.Itoa(id))
	snapshotPath := filepath.Join(migrationPath, "schema_snapshot.json")

	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema snapshot: %w", err)
	}

	var schema common.Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &schema, nil
}
