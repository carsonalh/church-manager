package main

import (
	"context"
	"log"
	"net/http"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connectionString := "postgres://postgres:admin@localhost:5432/churchmanager"

	err := PerformMigration("migrations", connectionString)
	if err != nil {
		log.Fatal(err.Error())
	}

	pgConnection, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		log.Fatalf("failed to connect to postgres database: %v", err)
	}

	memberHandler := CreateMemberHandler(CreateMemberPgStore(pgConnection), &MemberHandlerConfig{
		DefaultPageSize: 200,
		MaxPageSize:     500,
	})

	mux := http.NewServeMux()
	mux.Handle("/members", memberHandler)
	mux.Handle("/members/", memberHandler)

	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
