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
			FirstName:    NewPtr("Thomas"),
			LastName:     NewPtr("More"),
			EmailAddress: NewPtr("thomas.more.1478@gmail.com"),
			Notes:        NewPtr("Not to be put in the same Bible study as Luther"),
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

	t.Run("POST and GET /members index", func(t *testing.T) {
		client := TestRestClient{
			t:         t,
			serverUrl: server.URL,
		}

		member := Member{
			FirstName:    NewPtr("Martin"),
			LastName:     NewPtr("Luther"),
			EmailAddress: NewPtr("martin_luther@live.co.de"),
			PhoneNumber:  NewPtr("0428374598"),
		}

		var createdMember Member

		response := client.MakeRequest("POST", "/members", &member, &createdMember)
		if response.StatusCode != http.StatusCreated {
			t.Errorf("expected POST /members response to be 201 Created but was %s", response.Status)
		}
		if createdMember.Id == nil {
			t.Fatal("expected created member to have a valid id")
		}
		createdId := *createdMember.Id

		members := make([]Member, 0)

		response = client.MakeRequest("GET", "/members", nil, &members)
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected GET /members response to be 200 OK but was %s", response.Status)
		}

		recordFound := false
		for i, m := range members {
			if m.Id == nil {
				mJson, _ := json.Marshal(m)
				t.Errorf(
					"found a record with a nil id at index %d in the GET request, json: %s",
					i, string(mJson),
				)
			} else {
				if *m.Id == createdId {
					if recordFound {
						t.Error("found another record with the id being searched for")
					}
					recordFound = true
				}
			}
		}

		if !recordFound {
			t.Error("expected to find a record with the id of the created item in the index, but found none")
		}
	})

	t.Run("POST, DELETE and GET gives a 404", func(t *testing.T) {
		client := TestRestClient{
			t:         t,
			serverUrl: server.URL,
		}

		member := Member{
			FirstName: NewPtr("Carson"),
			LastName:  NewPtr("Holloway"),
		}

		response := client.MakeRequest("POST", "/members", &member, nil)
		location, err := response.Location()
		if err != nil {
			t.Fatal("error reading location from response")
		}

		response = client.MakeRequest("DELETE", location.Path, nil, nil)
		if response.StatusCode != http.StatusOK {
			t.Errorf(
				"expected response to have status 200 OK but got %s",
				response.Status,
			)
		}

		response = client.MakeRequest("GET", location.Path, nil, nil)

		if response.StatusCode != http.StatusNotFound {
			t.Errorf(
				"expected response to have status 404 Not Found but got %s",
				response.Status,
			)
		}
	})

	t.Run("POST, PUT and then GET returns updated data", func(t *testing.T) {
		client := TestRestClient{
			t:         t,
			serverUrl: server.URL,
		}

		member := Member{
			FirstName: NewPtr("John"),
			LastName:  NewPtr("Calvin"),
			Notes:     NewPtr("Still writing that very long book"),
		}

		response := client.MakeRequest("POST", "/members", &member, &member)
		location, err := response.Location()
		if err != nil {
			t.Fatalf("could not read Location header from response: %v", err)
		}

		member.Notes = NewPtr("Now re-writing the same book in French")

		response = client.MakeRequest("PUT", location.Path, &member, nil)
		if response.StatusCode != http.StatusOK {
			t.Errorf("PUT /members/{id} : expected status 200 OK but got %s", response.Status)
		}

		response = client.MakeRequest("GET", location.Path, nil, &member)
		if response.StatusCode != http.StatusOK {
			t.Errorf("GET /members/{id} : expected status 200 OK but got %s", response.Status)
		}

		if member.Notes == nil || *member.Notes != "Now re-writing the same book in French" {
			t.Error("updated data did not correctly persist accross calls")
		}
	})

	t.Run("POST, DELETE and DELETE returns a 404", func(t *testing.T) {
		client := TestRestClient{
			t:         t,
			serverUrl: server.URL,
		}

		member := Member{
			FirstName: NewPtr("John"),
			LastName:  NewPtr("Calvin"),
			Notes:     NewPtr("Still writing that very long book"),
		}

		response := client.MakeRequest("POST", "/members", &member, nil)
		location, err := response.Location()
		if err != nil {
			t.Fatalf("error reading location from response")
		}

		response = client.MakeRequest("DELETE", location.Path, nil, nil)
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected DELETE to be 200 OK, but was %s", response.Status)
		}

		response = client.MakeRequest("DELETE", location.Path, nil, nil)
		if response.StatusCode != http.StatusNotFound {
			t.Errorf("expected DELETE to be 404 Not Found, but was %s", response.Status)
		}
	})

	// TODO test pagination of data
}
