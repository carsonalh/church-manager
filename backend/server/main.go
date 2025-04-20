package main

import (
	"context"
	"log"

	_ "github.com/carsonalh/churchmanagerbackend/docs"
	"github.com/carsonalh/churchmanagerbackend/server/controller"
	"github.com/carsonalh/churchmanagerbackend/server/migration"
	"github.com/carsonalh/churchmanagerbackend/server/server"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Church Manager API
// @description     API for the Church Manager backend. Same api as used by the frontend.

// @host      localhost:8080
// @BasePath  /
func main() {
	connectionString := "postgres://postgres:admin@localhost:5432/churchmanager"

	err := migration.PerformMigration("migrations", connectionString)
	if err != nil {
		log.Fatal(err.Error())
	}

	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		log.Fatalf("failed to connect to postgres database: %v", err)
	}

	router := server.CreateServer(pool, server.ServerConfig{
		Members: controller.MemberControllerConfig{
			DefaultPageSize: 200,
			MaxPageSize:     500,
		},
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Fatal(router.Run("0.0.0.0:8080"))
}
