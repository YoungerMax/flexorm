package migrate

import (
	"database/sql"
	"fmt"

	"flexorm/common/v2"
)

// Config holds migration system configuration
type Config struct {
	MigrationsDir string
	SchemaFile    string
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		MigrationsDir: "migrations",
		SchemaFile:    "schema.json",
	}
}

// Migration represents a database migration
type Migration struct {
	ID             int
	UpSQL          string
	DownSQL        string
	SchemaSnapshot common.Schema
}

// Migrator handles all migration operations
type Migrator struct {
	config           Config
	sqlGenerator     SQLGenerator
	fileManager      FileManager
	migrationTracker MigrationTracker
}

// NewMigrator creates a new Migrator with default components
func NewMigrator(config Config, sqlGenerator SQLGenerator, fileManager FileManager, migrationTracker MigrationTracker) Migrator {
	return Migrator{
		config:           config,
		sqlGenerator:     sqlGenerator,
		fileManager:      fileManager,
		migrationTracker: migrationTracker,
	}
}

// GetAllMigrations returns all migrations sorted by ID
func (m *Migrator) ListMigrations() ([]Migration, error) {
	return m.fileManager.ListMigrations()
}

// GetMigration returns a specific migration by ID
func (m *Migrator) GetMigration(id int) (*Migration, error) {
	return m.fileManager.GetMigration(id)
}

func (m *Migrator) GetCurrentVersion(db *sql.DB) (int, error) {
	return m.migrationTracker.GetCurrentVersion(db)
}

// CreateMigrationOptions holds options for creating migrations
type CreateMigrationOptions struct {
	AutoGenerate bool
}

// CreateMigration creates a new migration with the given options
func (m *Migrator) CreateMigration(opts CreateMigrationOptions) (int, error) {
	// Get migrations
	migrations, err := m.ListMigrations()

	if err != nil {
		return 0, fmt.Errorf("failed to get migrations: %w", err)
	}

	// Determine next migration ID
	nextID := m.calculateNextID(migrations)

	// Load Schema
	currentSchema, err := common.LoadSchema(m.config.SchemaFile)

	if err != nil {
		return 0, err
	}

	// Create UpSQL and DownSQL
	var upSQL, downSQL string

	if opts.AutoGenerate {
		upSQL, downSQL, err = m.generateSQLFromSchemaDiff(*currentSchema)

		if err != nil {
			return 0, fmt.Errorf("failed to generate SQL: %w", err)
		}

		if upSQL == "" {
			return 0, fmt.Errorf("no schema changes detected")
		}
	} else {
		upSQL = "-- Write your up migration here\n\n"
		downSQL = "-- Write your down migration here\n\n"
	}

	// Create migration files
	migration := Migration{
		ID:             nextID,
		UpSQL:          upSQL,
		DownSQL:        downSQL,
		SchemaSnapshot: *currentSchema,
	}

	if err := m.fileManager.CreateMigration(migration); err != nil {
		return 0, fmt.Errorf("failed to create migration files: %w", err)
	}

	return nextID, nil
}

// MigrateUpOptions holds options for migrating up
type MigrateUpOptions struct {
	TargetID int
}

// MigrateUp runs migrations up to the specified ID (or all if nil)
func (m *Migrator) MigrateUp(db *sql.DB, opts MigrateUpOptions) error {
	// Initialize migration tracking
	if err := m.migrationTracker.Initialize(db); err != nil {
		return fmt.Errorf("failed to initialize migration tracking: %w", err)
	}

	// Get current version
	currentVersion, err := m.migrationTracker.GetCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Get all migrations
	migrations, err := m.ListMigrations()
	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}

	// Filter migrations to run
	toRun := m.filterMigrationsToRun(migrations, currentVersion, opts.TargetID)

	if len(toRun) == 0 {
		fmt.Println("No migrations to run")

		return nil
	}

	// Execute migrations
	for _, migration := range toRun {
		fmt.Printf("Running migration #%d\n", migration.ID)

		if err := m.executeMigrationUp(db, migration); err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", migration.ID, err)
		}

		fmt.Printf("Completed migration #%d\n", migration.ID)
	}

	return nil
}

// MigrateDownOptions holds options for migrating down
type MigrateDownOptions struct {
	TargetID int
}

// MigrateDown rolls back migrations
func (m *Migrator) MigrateDown(db *sql.DB, opts MigrateDownOptions) error {
	// Initialize migration tracking
	if err := m.migrationTracker.Initialize(db); err != nil {
		return fmt.Errorf("failed to initialize migration tracking: %w", err)
	}

	// Get current version
	currentVersion, err := m.migrationTracker.GetCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion < 0 {
		fmt.Println("No migrations to roll back")

		return nil
	}

	// Get all migrations
	migrations, err := m.ListMigrations()
	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}

	// Filter migrations to rollback
	toRollback := m.filterMigrationsToRollback(migrations, currentVersion, opts.TargetID)

	if len(toRollback) == 0 {
		fmt.Println("No migrations to roll back")
		return nil
	}

	// Execute rollbacks
	for _, migration := range toRollback {
		fmt.Printf("Rolling back migration #%d\n", migration.ID)

		if err := m.executeMigrationDown(db, migration); err != nil {
			return fmt.Errorf("failed to rollback migration %d: %w", migration.ID, err)
		}

		fmt.Printf("Rolled back migration #%d\n", migration.ID)
	}

	return nil
}

// Private helper methods
func (m *Migrator) calculateNextID(migrations []Migration) int {
	if len(migrations) == 0 {
		return 0
	}
	return migrations[len(migrations)-1].ID + 1
}

func (m *Migrator) generateSQLFromSchemaDiff(currentSchema common.Schema) (string, string, error) {
	// Get previous schema
	var previousSchema common.Schema
	migrations, err := m.ListMigrations()

	if err != nil {
		return "", "", err
	}

	if len(migrations) == 0 {
		previousSchema = common.EmptySchema
	} else {
		previousSchema = migrations[len(migrations)-1].SchemaSnapshot
	}

	// Generate SQL diff
	diff := GenerateDiff(previousSchema, currentSchema)
	upSQL, downSQL := m.sqlGenerator.GenerateSQL(diff)

	return upSQL, downSQL, nil
}

func (m *Migrator) executeMigrationUp(db *sql.DB, migration Migration) error {
	// Execute SQL
	if _, err := db.Exec(migration.UpSQL); err != nil {
		return err
	}

	// Record migration
	return m.migrationTracker.RecordMigration(db, migration.ID)
}

func (m *Migrator) executeMigrationDown(db *sql.DB, migration Migration) error {
	// Execute SQL
	if _, err := db.Exec(migration.DownSQL); err != nil {
		return err
	}

	// Remove migration record
	return m.migrationTracker.RemoveMigration(db, migration.ID)
}

func (m *Migrator) filterMigrationsToRun(migrations []Migration, currentVersion, targetVersion int) []Migration {
	var result []Migration
	for _, mig := range migrations {
		if mig.ID > currentVersion && mig.ID <= targetVersion {
			result = append(result, mig)
		}
	}
	return result
}

func (m *Migrator) filterMigrationsToRollback(migrations []Migration, currentVersion, targetVersion int) []Migration {
	var result []Migration
	for i := len(migrations) - 1; i >= 0; i-- {
		mig := migrations[i]
		if mig.ID <= currentVersion && mig.ID > targetVersion {
			result = append(result, mig)
		}
	}
	return result
}
