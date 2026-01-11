package migrate

import (
	"database/sql"
	"fmt"
)

// SQLMigrationTracker tracks migrations in a SQL database
type SQLMigrationTracker struct {
	tableName string
}

// NewSQLMigrationTracker creates a new SQL migration tracker
func NewSQLMigrationTracker(tableName string) *SQLMigrationTracker {
	if tableName == "" {
		tableName = "schema_migrations"
	}
	return &SQLMigrationTracker{
		tableName: tableName,
	}
}

// Initialize creates the migrations tracking table
func (t *SQLMigrationTracker) Initialize(db *sql.DB) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			description TEXT
		);
	`, t.tableName)

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// GetCurrentVersion returns the highest applied migration version
func (t *SQLMigrationTracker) GetCurrentVersion(db *sql.DB) (int, error) {
	query := fmt.Sprintf("SELECT COALESCE(MAX(version), -1) FROM %s;", t.tableName)

	var version int
	err := db.QueryRow(query).Scan(&version)
	if err != nil {
		return -1, fmt.Errorf("failed to get current version: %w", err)
	}

	return version, nil
}

// GetAppliedMigrations returns all applied migration versions
func (t *SQLMigrationTracker) GetAppliedMigrations(db *sql.DB) ([]int, error) {
	query := fmt.Sprintf("SELECT version FROM %s ORDER BY version;", t.tableName)

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, version)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return versions, nil
}

// RecordMigration records that a migration has been applied
func (t *SQLMigrationTracker) RecordMigration(db *sql.DB, version int) error {
	query := fmt.Sprintf("INSERT INTO %s (version) VALUES ($1);", t.tableName)

	_, err := db.Exec(query, version)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

// RemoveMigration removes a migration record
func (t *SQLMigrationTracker) RemoveMigration(db *sql.DB, version int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE version = $1;", t.tableName)

	_, err := db.Exec(query, version)
	if err != nil {
		return fmt.Errorf("failed to remove migration: %w", err)
	}

	return nil
}
