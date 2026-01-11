package main

import (
	"database/sql"
	"flexorm/clientgen/v2"
	"flexorm/common/v2"
	"flexorm/migrate/v2"
	"flexorm/studio/v2"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	schemaFile     string
	migrationsDir  string
	databaseURL    string
	databaseDriver string

	// Migration flags
	autoGenerate  bool
	targetVersion int
	stepCount     int

	// Studio flags
	studioHost string
	studioPort int
)

var rootCmd = &cobra.Command{
	Use:          "flexorm",
	Short:        "The last ORM you'll ever need",
	Long:         `FlexORM is a powerful ORM that generates type-safe database clients and manages schema migrations with ease.`,
	SilenceUsage: true,
}

var generateCmd = &cobra.Command{
	Use:   "generate [target]",
	Short: "Generate a type-safe database client",
	Long:  `Generates a type-safe database client for a given language-driver target from your schema.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  generateTask,
}

var migrationsCmd = &cobra.Command{
	Use:   "migrations",
	Short: "Manage database migrations",
	Long:  `Commands for creating and running database migrations.`,
}

var migrationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new migration",
	Long: `Create a new migration file.

With --auto flag, generates migration by comparing schema.json with the last migration.
Without --auto flag, creates an empty migration template for manual editing.`,
	Args: cobra.MaximumNArgs(1),
	RunE: migrateCreateTask,
}

var migrationsUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run pending migrations",
	Long: `Run all pending migrations or up to a specific version.

Examples:
  flexorm migrations up --to 5             # Run migrations up to version 5`,
	RunE: migrateUpTask,
}

var migrationsDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback migrations",
	Long: `Rollback migrations to a previous version.

Examples:
  flexorm migrations down --to 0           # Rollback all migrations`,
	RunE: migrateDownTask,
}

var studioCmd = &cobra.Command{
	Use:   "studio",
	Short: "Launch FlexORM Studio web interface",
	Long:  `Studio launches a web-based database explorer to view and manage your database.`,
	RunE:  studioTask,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&schemaFile, "schema", "schema.json", "Path to schema file")
	rootCmd.PersistentFlags().StringVar(&migrationsDir, "migrations-dir", "migrations", "Path to migrations directory")
	rootCmd.PersistentFlags().StringVar(&databaseURL, "database-url", "", "Database connection URL")
	rootCmd.PersistentFlags().StringVar(&databaseDriver, "driver", "postgres", "Database driver (postgres, sqlite)")

	// Compile command flags
	generateCmd.Flags().StringP("output", "o", "-", "Output file for generated code or - or stdout")

	// Migration create flags
	migrationsCreateCmd.Flags().BoolVar(&autoGenerate, "auto", true, "Auto-generate migration from schema diff")
	migrationsCreateCmd.Flags().Bool("empty", false, "Create empty migration template")

	// Migration up flags
	migrationsUpCmd.Flags().IntVar(&targetVersion, "to", -1, "Target migration version")
	migrationsUpCmd.Flags().IntVar(&stepCount, "step", 0, "Number of migrations to run")

	// Migration down flags
	migrationsDownCmd.Flags().IntVar(&targetVersion, "to", -1, "Target migration version (use 0 for complete rollback)")
	migrationsDownCmd.Flags().IntVar(&stepCount, "step", 1, "Number of migrations to rollback")

	// Studio command flags
	studioCmd.Flags().StringVar(&studioHost, "host", "localhost", "Host to bind the studio server")
	studioCmd.Flags().IntVar(&studioPort, "port", 5555, "Port to bind the studio server")

	// Add subcommands
	migrationsCmd.AddCommand(
		migrationsCreateCmd,
		migrationsUpCmd,
		migrationsDownCmd,
	)

	rootCmd.AddCommand(
		generateCmd,
		migrationsCmd,
		studioCmd,
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// generateTask generates a type-safe database client
func generateTask(cmd *cobra.Command, args []string) error {
	schema, err := common.LoadSchema(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	target := clientgen.TypeScriptPostgresJS
	if len(args) > 0 {
		target = clientgen.Target(args[0])
	}

	output, _ := cmd.Flags().GetString("output")

	if output == "-" {
		if err := clientgen.Emit(*schema, target, os.Stdout); err != nil {
			return fmt.Errorf("client generation failed: %w", err)
		}
	} else {
		fmt.Printf("Generating client for to %s...\n", target)

		fp, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}

		if err := clientgen.Emit(*schema, target, fp); err != nil {
			return fmt.Errorf("client generation failed: %w", err)
		}

		fmt.Printf("✓ Successfully generated client for %s\n", output)
	}

	return nil
}

// migrateCreateTask creates a new migration
func migrateCreateTask(cmd *cobra.Command, args []string) error {
	migrator := createMigrator()

	empty, _ := cmd.Flags().GetBool("empty")
	opts := migrate.CreateMigrationOptions{
		AutoGenerate: autoGenerate && !empty,
	}

	id, err := migrator.CreateMigration(opts)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	fmt.Printf("✓ Created migration %d in %s/%d/\n", id, migrationsDir, id)

	if opts.AutoGenerate {
		fmt.Printf("  Migration was auto-generated from schema changes\n")
	} else {
		fmt.Printf("  Edit up.sql and down.sql to define your migration\n")
	}

	return nil
}

// migrateUpTask runs pending migrations
func migrateUpTask(cmd *cobra.Command, args []string) error {
	db, err := connectDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	migrator := createMigrator()

	opts := migrate.MigrateUpOptions{
		TargetID: targetVersion,
	}

	if err := migrator.MigrateUp(db, opts); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("✓ Migrations completed successfully")
	return nil
}

// migrateDownTask rolls back migrations
func migrateDownTask(cmd *cobra.Command, args []string) error {
	db, err := connectDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	migrator := createMigrator()

	opts := migrate.MigrateDownOptions{
		TargetID: targetVersion,
	}

	if err := migrator.MigrateDown(db, opts); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Println("✓ Rollback completed successfully")
	return nil
}

// studioTask launches the FlexORM Studio web interface
func studioTask(cmd *cobra.Command, args []string) error {
	if databaseURL == "" {
		return fmt.Errorf("database URL required (use --database-url flag)")
	}

	return studio.RunWebServer(studioHost, studioPort, databaseURL, databaseDriver)
}

// Helper functions

func createMigrator() *migrate.Migrator {
	config := migrate.Config{
		MigrationsDir: migrationsDir,
		SchemaFile:    schemaFile,
	}

	sqlGen := createSQLGenerator()
	fileManager := migrate.LocalFileManager{Config: config}
	tracker := createMigrationTracker()

	migrator := migrate.NewMigrator(config, sqlGen, fileManager, tracker)
	return &migrator
}

func createSQLGenerator() migrate.SQLGenerator {
	switch databaseDriver {
	case "postgres", "postgresql":
		return migrate.PostgreSQLGenerator{}
	default:
		return migrate.PostgreSQLGenerator{}
	}
}

func createMigrationTracker() migrate.MigrationTracker {
	return migrate.NewSQLMigrationTracker("schema_migrations")
}

func connectDatabase() (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL required (use --database-url flag)")
	}

	db, err := sql.Open(databaseDriver, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
