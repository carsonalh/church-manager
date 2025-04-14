package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Member struct {
	Id        *uint64 `json:"id"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

type MemberHandler struct {
	mux   *http.ServeMux
	store *MemberPgStore
}

func (h *MemberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// Member handler is a simple crud route
func CreateMemberHandler(store *MemberPgStore) *MemberHandler {
	mux := http.NewServeMux()
	handler := &MemberHandler{mux: mux, store: store}

	mux.HandleFunc("GET /members", handler.getMembers)
	mux.HandleFunc("GET /members/{id}", handler.getMember)
	mux.HandleFunc("POST /members", handler.postMember)

	return handler
}

func (h *MemberHandler) getMembers(w http.ResponseWriter, r *http.Request) {
	var members []Member
	var err error

	if members, err = h.store.Get(500, 0); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(members)
}

func (h *MemberHandler) postMember(w http.ResponseWriter, r *http.Request) {
	var err error
	var body Member

	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.Id != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	insertErr := h.store.Insert(&body)
	if insertErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/members/%d", *body.Id))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (h *MemberHandler) getMember(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	member, err := h.store.FindById(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if member == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(*member)
}
