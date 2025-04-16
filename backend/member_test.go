package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupTestSuite(tb testing.TB) func(testing.TB) {
	return func(tb testing.TB) {
	}
}

type TestRestClient struct {
	t         *testing.T
	serverUrl string
}

func (c *TestRestClient) MakeRequest(method string, url string, body any, responseBody any) *http.Response {
	requestData := make([]byte, 0)

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			c.t.Fatalf("%s %s : failed to encode body to json data, body: %v", method, url, body)
		}
		requestData = jsonData
	}

	request, err := http.NewRequest(method, c.serverUrl+url, bytes.NewReader(requestData))
	if err != nil {
		c.t.Fatalf("%s %s : failed to create http request: %v", method, url, err)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		c.t.Fatalf("%s %s : failed to send http request: %v", method, url, err)
	}

	if responseBody != nil {
		err = json.NewDecoder(response.Body).Decode(responseBody)
		if err != nil {
			c.t.Fatalf("%s %s : unable to read response json into object %v", method, url, responseBody)
		}
	}

	return response
}

func TestMemberRest(t *testing.T) {
	connectionString := "postgres://postgres:admin@localhost:5432/churchmanager"
	err := PerformMigration(connectionString)
	if err != nil {
		t.Fatalf("database migration error: %v", err)
	}
	conn, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		t.Fatalf("could not connect to the database")
	}
	defer conn.Close()

	server := httptest.NewServer(CreateMemberHandler(CreateMemberPgStore(conn)))
	defer server.Close()

	t.Run("POST and GET again", func(t *testing.T) {
		client := TestRestClient{
			t:         t,
			serverUrl: server.URL,
		}

		requestBody := Member{
			FirstName:    NewInit("Thomas"),
			LastName:     NewInit("More"),
			EmailAddress: NewInit("thomas.more.1478@gmail.com"),
			Notes:        NewInit("Not to be put in the same Bible study as Luther"),
		}

		firstResponse := Member{}

		response := client.MakeRequest("POST", "/members", &requestBody, &firstResponse)
		if firstResponse.Id == nil {
			t.Errorf("response body should have an id")
		}
		if response.StatusCode != http.StatusCreated {
			t.Errorf("expected status 201 Created, but got %s", response.Status)
		}
		location, err := response.Location()
		if err != nil {
			t.Fatalf("could not read Location header from response: %v", err)
		}

		finalResponse := Member{}
		_ = client.MakeRequest("GET", location.Path, nil, &finalResponse)

		derefOrNil := func(x *string) any {
			if x == nil {
				return nil
			}
			return *x
		}

		if derefOrNil(requestBody.FirstName) != derefOrNil(finalResponse.FirstName) ||
			derefOrNil(requestBody.LastName) != derefOrNil(finalResponse.LastName) ||
			derefOrNil(requestBody.EmailAddress) != derefOrNil(finalResponse.EmailAddress) ||
			derefOrNil(requestBody.Notes) != derefOrNil(finalResponse.Notes) {
			requestJson, _ := json.Marshal(requestBody)
			responseJson, _ := json.Marshal(finalResponse)
			t.Errorf(
				"Member object was not correctly reproduced by the server, "+
					"expected %v and got %v (excluding Id field)",
				string(requestJson),
				string(responseJson),
			)
		}
	})

	t.Logf("finished running sub-test")
}
