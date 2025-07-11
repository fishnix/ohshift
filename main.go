// Package main is the entry point for the OhShift! Slack incident management bot.
package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib" // import the postgres driver
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"

	dbm "github.com/fishnix/ohshift/db"
	"github.com/fishnix/ohshift/internal/config"
	"github.com/fishnix/ohshift/internal/logger"
	"github.com/fishnix/ohshift/internal/slack"
)

var cfg *config.Config

func main() {
	rootCmd := &cobra.Command{
		Use:   "ohshift",
		Short: "OhShift! Slack incident management bot",
		Long:  `OhShift! is a Slack bot for managing incidents and on-call rotations.`,
	}

	botCmd := &cobra.Command{
		Use:   "bot",
		Short: "Start the Slack bot",
		Long:  `Start the OhShift! Slack bot for incident management.`,
		RunE:  runBot,
	}

	migrateCmd := &cobra.Command{
		Use:   "migrate <command> [args]",
		Short: "Migrate the database",
		Long: `Migrate provides a wrapper around the "goose" migration tool.

	Commands:
	up                   Migrate the DB to the most recent version available
	up-by-one            Migrate the DB up by 1
	up-to VERSION        Migrate the DB to a specific VERSION
	down                 Roll back the version by 1
	down-to VERSION      Roll back to a specific VERSION
	redo                 Re-run the latest migration
	reset                Roll back all migrations
	status               Dump the migration status for the current DB
	version              Print the current version of the database
	create NAME [sql|go] Creates new migration file with the current timestamp
	fix                  Apply sequential ordering to migrations
	`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runMigration(cmd.Context(), args[0], args[1:]); err != nil {
				logger.Fatal("Migration failed", "error", err)
			}
		},
	}

	rootCmd.AddCommand(botCmd, migrateCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runBot(_ *cobra.Command, _ []string) error {
	// Load configuration
	cfg = config.Load()

	// Set log level from configuration
	logger.SetLevel(cfg.LogLevel)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatal("Configuration error", "error", err)
		return err
	}

	db := initDB()
	runMigrationInternal(db.DB)

	// Create Slack bot
	bot := slack.NewBot(cfg)

	// Set up context with cancel on SIGINT/SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start the bot (blocks until shutdown)
	if err := bot.Start(ctx); err != nil {
		logger.Fatal("Bot error", "error", err)
		return err
	}

	logger.Info("Bot exited cleanly")

	return nil
}

func runMigration(ctx context.Context, command string, args []string) error {
	// Load configuration
	cfg = config.Load()

	// Set log level from configuration
	logger.SetLevel(cfg.LogLevel)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatal("Configuration error", "error", err)
		return err
	}

	db, err := goose.OpenDBWithDriver("postgres", cfg.DBURI)
	if err != nil {
		logger.Fatal("failed to open DB", "error", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal("failed to close DB", "error", err)
		}
	}()

	goose.SetBaseFS(dbm.Migrations)

	if err := goose.RunContext(ctx, command, db, "migrations", args...); err != nil {
		logger.Fatal("migrate command failed", "command", command, "error", err)
		return err
	}

	logger.Info("Migrations completed successfully")

	return nil
}

// runMigrationInternal is the internal migration function used by both bot and migrate commands
func runMigrationInternal(db *sql.DB) {
	goose.SetBaseFS(dbm.Migrations)

	if err := goose.Up(db, "migrations"); err != nil {
		logger.Fatal("migration failed", "error", err)
	}
}

func initDB() *sqlx.DB {
	dbDriverName := "postgres"

	connector, err := pq.NewConnector(cfg.DBURI)
	if err != nil {
		logger.Fatal("failed initializing sql connector", "error", err)
	}

	db := sqlx.NewDb(sql.OpenDB(connector), dbDriverName)

	if err := db.PingContext(context.Background()); err != nil {
		logger.Fatal("failed verifying database connection", "error", err)
	}

	// TODO: configure connection pool
	// db.SetMaxOpenConns(viper.GetInt("db.connections.max_open"))
	// db.SetMaxIdleConns(viper.GetInt("db.connections.max_idle"))
	// db.SetConnMaxIdleTime(viper.GetDuration("db.connections.max_lifetime"))

	return db
}
