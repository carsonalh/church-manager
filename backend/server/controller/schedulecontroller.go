package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/carsonalh/churchmanagerbackend/server/domain"
	"github.com/carsonalh/churchmanagerbackend/server/store"
	"github.com/gin-gonic/gin"
)

type ScheduleHandler struct {
	store *store.ScheduleStore
}

func SetupScheduleHandler(router *gin.RouterGroup, store *store.ScheduleStore) {
	handler := ScheduleHandler{store: store}

	router.POST("", handler.postSchedule)
}

func (h *ScheduleHandler) postSchedule(c *gin.Context) {
	var createDto domain.ScheduleCreateDTO

	if err := c.BindJSON(&createDto); err != nil {
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		var timeParseErr *time.ParseError
		switch {
		case errors.As(err, &syntaxErr):
			c.String(http.StatusBadRequest, fmt.Sprintf("syntax error at character %d\n", syntaxErr.Offset))
		case errors.As(err, &typeErr):
			c.String(http.StatusBadRequest, fmt.Sprintf("type error for field %s\n", typeErr.Field))
		case errors.As(err, &timeParseErr):
			c.String(http.StatusBadRequest, fmt.Sprintf("error parsing timestamp: %s\n", timeParseErr.Error()))
		}
		return
	}

	errs := createDto.Validate()
	if len(errs) != 0 {
		c.JSON(http.StatusBadRequest, errs)
		return
	}

	schedule, err := h.store.Create(&createDto)
	if err != nil {
		log.Printf("error inserting into database: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, schedule.ToResponseDTO())
}
