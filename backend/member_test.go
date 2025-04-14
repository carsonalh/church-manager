package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestCrudOperationsThroughHTTP(t *testing.T) {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:admin@localhost:5432/churchmanager")
	if err != nil {
		t.Fatalf("could not connect to the database")
	}
	defer conn.Close(context.Background())

	server := httptest.NewServer(CreateMemberHandler(CreateMemberPgStore(conn)))
	defer server.Close()

	var members []Member
	response, err := http.Get(server.URL + "/members")
	if err != nil {
		t.Fatalf("error making GET request to server")
	}
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("error reading body from server response")
	}
	err = json.Unmarshal(bytes, &members)
	if err != nil {
		s := string(bytes)
		t.Fatalf("server response was not valid json, response: %s", s)
	}

	if len(members) != 0 {
		t.Errorf("expected to start with an empty data store")
	}
}
