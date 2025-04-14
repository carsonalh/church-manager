package main

import (
	"log"
	"net/http"
)

func main() {
	memberHandler := CreateMemberHandler(CreateMemberPgStore())

	mux := http.NewServeMux()
	mux.Handle("/members", memberHandler)
	mux.Handle("/members/", memberHandler)

	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
