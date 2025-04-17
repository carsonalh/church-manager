package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Member struct {
	Id           *uint64 `json:"id"`
	FirstName    *string `json:"firstName"`
	LastName     *string `json:"lastName"`
	EmailAddress *string `json:"emailAddress"`
	PhoneNumber  *string `json:"phoneNumber"`
	Notes        *string `json:"notes"`
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
	mux.HandleFunc("POST /members", handler.postMember)
	mux.HandleFunc("GET /members/{id}", handler.getMember)
	mux.HandleFunc("PUT /members/{id}", handler.putMember)
	mux.HandleFunc("DELETE /members/{id}", handler.deleteMember)

	return handler
}

func (h *MemberHandler) getMembers(w http.ResponseWriter, r *http.Request) {
	var members []Member
	var err error

	if members, err = h.store.Get(500, 0); err != nil {
		log.Printf("GET /members : error getting members from database: %v", err)
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
		if body.Id != nil {
			_, _ = io.WriteString(w, "new member must not have id field")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, insertErr := h.store.Insert(&body)
	if insertErr != nil {
		log.Printf("POST /members : error inserting into database: %v", insertErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/members/%d", *result.Id))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *MemberHandler) getMember(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	member, err := h.store.FindById(id)
	if err != nil {
		log.Printf("GET /member/{id} : %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if member == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(*member)
	if err != nil {
		log.Printf("failed to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *MemberHandler) putMember(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var member Member
	err = json.NewDecoder(r.Body).Decode(&member)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.store.Update(id, &member)
	if err != nil {
		log.Printf("PUT /members/{id} : error updating member: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	member.Id = NewPtr(id)
	err = json.NewEncoder(w).Encode(member)
	if err != nil {
		log.Printf("PUT /members/{id} : failed to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *MemberHandler) deleteMember(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	deleted, err := h.store.Delete(id)
	if err != nil {
		log.Printf("DELETE /member/{id} : %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if deleted {
		w.WriteHeader(http.StatusOK)
	} else {
		// The only conceivable reason why a delete count would be zero
		w.WriteHeader(http.StatusNotFound)
	}
}
