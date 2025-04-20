package server

import (
	"github.com/carsonalh/churchmanagerbackend/server/controller"
	"github.com/carsonalh/churchmanagerbackend/server/store"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerConfig struct {
	Schedules struct{}
	Members   controller.MemberControllerConfig
}

func CreateServer(pool *pgxpool.Pool, config ServerConfig) *gin.Engine {
	router := gin.Default()

	controller.SetupScheduleHandler(router.Group("/schedules"), store.CreateScheduleStore(pool))
	controller.SetupMemberController(router.Group("/members"), store.CreateMemberStore(pool), &controller.MemberControllerConfig{
		DefaultPageSize: config.Members.DefaultPageSize,
		MaxPageSize:     config.Members.MaxPageSize,
	})

	return router
}
