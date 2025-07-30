package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func RunMigrations(connInfo Connection, migrationsPath string) error {
	log.Println("Starting database migrations...")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		connInfo.Host, connInfo.Port, connInfo.User, connInfo.Password, connInfo.DBName, connInfo.SSLMode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("failed to open database connection for migrations: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	currentVersion, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	if dirty {
		log.Printf("Warning: Database is in dirty state at version %d", currentVersion)
		return fmt.Errorf("database is in dirty state at version %d, manual intervention required", currentVersion)
	}

	if err == migrate.ErrNilVersion {
		log.Println("No migrations have been applied yet")
	} else {
		log.Printf("Current migration version: %d", currentVersion)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	newVersion, _, err := m.Version()
	if err != nil {
		log.Println("Migrations applied successfully, but failed to get new version")
	} else {
		log.Printf("Migrations applied successfully. New version: %d", newVersion)
	}

	return nil
}
