package main

import (
	"context"
	"log"

	"github.com/carsonalh/churchmanagerbackend/server/controller"
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

	router := CreateServer(pool, ServerConfig{
		Members: controller.MemberControllerConfig{
			DefaultPageSize: 200,
			MaxPageSize:     500,
		},
	})

	log.Fatal(router.Run("0.0.0.0:8080"))
}
