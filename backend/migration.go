package main

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
)

// Creates a database connection, migrates the database to the most current
// version, and destroys the connection.
func PerformMigration(connectionString string) error {
	migration, err := migrate.New("file://migrations", connectionString+"?sslmode=disable")

	if err != nil {
		return fmt.Errorf("migration client failed to initialise: %v", err)
	}
	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %v", err)
	}
	fileErr, dbErr := migration.Close()
	if fileErr != nil {
		return fmt.Errorf("failed to close files after migration: %v", fileErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close database after migration: %v", dbErr)
	}

	return nil
}
