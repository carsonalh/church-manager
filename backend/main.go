package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
)

func main() {
	connectionString := "postgres://postgres:admin@localhost:5432/churchmanager"

	migration, err := migrate.New("file://migrations", connectionString+"?sslmode=disable")
	if err != nil {
		log.Fatalf("migration client failed to initialise: %v", err)
	}
	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration failed: %v", err)
	}
	fileErr, dbErr := migration.Close()
	if fileErr != nil {
		log.Fatalf("failed to close files after migration: %v", fileErr)
	}
	if dbErr != nil {
		log.Fatalf("failed to close database after migration: %v", dbErr)
	}

	pgConnection, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		log.Fatalf("failed to connect to postgres database: %v", err)
	}

	memberHandler := CreateMemberHandler(CreateMemberPgStore(pgConnection))

	mux := http.NewServeMux()
	mux.Handle("/members", memberHandler)
	mux.Handle("/members/", memberHandler)

	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
