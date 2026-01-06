package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/blagoySimandov/ampledata/go/internal/config"
	"github.com/blagoySimandov/ampledata/go/migrations"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
)

func main() {
	cfg := config.Load()

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.DatabaseURL)))
	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()

	migrator := migrate.NewMigrator(db, migrations.Migrations)

	ctx := context.Background()

	if err := migrator.Init(ctx); err != nil {
		log.Fatalf("Failed to initialize migrator: %v", err)
	}

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "up":
		group, err := migrator.Migrate(ctx)
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		if group.IsZero() {
			fmt.Println("No new migrations to run (database is up to date)")
			return
		}
		fmt.Printf("Migrated to %s\n", group)

	case "down":
		group, err := migrator.Rollback(ctx)
		if err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		if group.IsZero() {
			fmt.Println("No migrations to rollback")
			return
		}
		fmt.Printf("Rolled back %s\n", group)

	case "status":
		ms, err := migrator.MigrationsWithStatus(ctx)
		if err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		fmt.Printf("Migrations:\n")
		for _, m := range ms {
			status := "pending"
			if m.IsApplied() {
				status = "applied"
			}
			fmt.Printf("  %s: %s\n", m.Name, status)
		}

	case "create":
		name := "migration"
		if len(os.Args) > 2 {
			name = strings.Join(os.Args[2:], "_")
		}
		files, err := migrator.CreateTxSQLMigrations(ctx, name)
		if err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
		for _, f := range files {
			fmt.Printf("Created migration: %s\n", f.Path)
		}

	default:
		fmt.Println("Usage: migrate [up|down|status|create <name>]")
		fmt.Println("  up     - Run all pending migrations")
		fmt.Println("  down   - Rollback the last migration group")
		fmt.Println("  status - Show migration status")
		fmt.Println("  create - Create new migration files")
		os.Exit(1)
	}
}
