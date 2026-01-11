package migrate

import (
	"database/sql"
	"flexorm/common/v2"
)

// SQLGenerator interface for database-specific SQL generation
type SQLGenerator interface {
	GenerateSQL(diff *SchemaDiff) (upSQL, downSQL string)
}

// FileManager manages migration files on the filesystem
type FileManager interface {
	CreateMigration(migration Migration) error
	GetMigration(id int) (*Migration, error)
	ListMigrations() ([]Migration, error)
	SaveSchemaSnapshot(id int, schema common.Schema) error
	LoadSchemaSnapshot(id int) (*common.Schema, error)
}

// MigrationTracker tracks which migrations have been applied
type MigrationTracker interface {
	Initialize(db *sql.DB) error
	GetCurrentVersion(db *sql.DB) (int, error)
	GetAppliedMigrations(db *sql.DB) ([]int, error)
	RecordMigration(db *sql.DB, version int) error
	RemoveMigration(db *sql.DB, version int) error
}

// ChangeType represents the type of schema change
type ChangeType string

const (
	ChangeTypeCreateTable ChangeType = "create_table"
	ChangeTypeDropTable   ChangeType = "drop_table"
	ChangeTypeAddColumn   ChangeType = "add_column"
	ChangeTypeDropColumn  ChangeType = "drop_column"
	ChangeTypeAlterColumn ChangeType = "alter_column"
)

// SchemaChange represents a single schema change in the AST
type SchemaChange struct {
	Type      ChangeType
	TableName string
	Details   interface{}
}

// TableCreate represents a table creation change
type TableCreate struct {
	Table common.Table
}

// TableDrop represents a table drop change
type TableDrop struct {
	Table common.Table
}

// ColumnAdd represents an added column
type ColumnAdd struct {
	ColumnName string
	Column     common.Column
}

// ColumnDrop represents a dropped column
type ColumnDrop struct {
	ColumnName string
	Column     common.Column
}

// ColumnAlter represents a modified column
type ColumnAlter struct {
	ColumnName string
	OldColumn  common.Column
	NewColumn  common.Column
}

// SchemaDiff represents the difference between two schemas
type SchemaDiff struct {
	Changes []SchemaChange
}
