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
)

type ScheduleHandler struct {
	mux   *http.ServeMux
	store *store.ScheduleStore
}

func CreateScheduleHandler(store *store.ScheduleStore) *ScheduleHandler {
	mux := http.NewServeMux()

	handler := ScheduleHandler{
		mux:   mux,
		store: store,
	}

	mux.HandleFunc("POST /schedules", handler.postChurchServiceSchedule)

	return &handler
}

func (h *ScheduleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *ScheduleHandler) postChurchServiceSchedule(w http.ResponseWriter, r *http.Request) {
	var createDto domain.ScheduleCreateDTO

	err := json.NewDecoder(r.Body).Decode(&createDto)
	if err != nil {
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		var timeParseErr *time.ParseError
		w.WriteHeader(http.StatusBadRequest)
		switch {
		case errors.As(err, &syntaxErr):
			w.Write([]byte(fmt.Sprintf("syntax error at character %d\n", syntaxErr.Offset)))
		case errors.As(err, &typeErr):
			w.Write([]byte(fmt.Sprintf("type error for field %s\n", typeErr.Field)))
		case errors.As(err, &timeParseErr):
			w.Write([]byte(fmt.Sprintf("error parsing timestamp: %s\n", timeParseErr.Error())))
		}
		return
	}

	errs := createDto.Validate()
	if len(errs) != 0 {
		w.WriteHeader(http.StatusBadRequest)
		// TODO display errors to user
		return
	}

	schedule, err := h.store.Create(&createDto)
	if err != nil {
		log.Printf("error inserting into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(schedule.ToResponseDTO())
}
