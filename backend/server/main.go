package main

import (
	"context"
	"log"
	"net/http"

	"github.com/carsonalh/churchmanagerbackend/server/controller"
	"github.com/carsonalh/churchmanagerbackend/server/store"
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

	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		log.Fatalf("failed to connect to postgres database: %v", err)
	}

	memberHandler := CreateMemberHandler(CreateMemberPostgresStore(pool), &MemberHandlerConfig{
		DefaultPageSize: 200,
		MaxPageSize:     500,
	})

	churchServiceHandler := controller.CreateScheduleHandler(store.CreateScheduleStore(pool))

	mux := http.NewServeMux()
	mux.Handle("/members", memberHandler)
	mux.Handle("/members/", memberHandler)
	mux.Handle("/schedules", churchServiceHandler)
	mux.Handle("/schedules/", churchServiceHandler)

	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
