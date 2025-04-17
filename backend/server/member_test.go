package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

type TestPostgresContainer struct {
	container testcontainers.Container
	logs      io.ReadCloser
	logFile   *os.File
}

func CreateTestContainer(tb testing.TB) (container *TestPostgresContainer, connectionString string) {
	user := "postgres"
	password := "admin"
	database := "churchmanager"

	container = new(TestPostgresContainer)

	var err, containerErr error
	container.container, containerErr = testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:17.4",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       database,
				"POSTGRES_USER":     user,
				"POSTGRES_PASSWORD": password,
			},
			WaitingFor: wait.ForExposedPort(),
		},
		Started: true,
	})
	testcontainers.CleanupContainer(tb, container.container)
	container.logs, err = container.container.Logs(context.Background())
	if err != nil {
		tb.Fatalf("error getting logs from test container: %v", err)
	}
	tb.Cleanup(func() { container.logs.Close() })
	_ = os.Mkdir("out", 0o755) // rwxr-xr-x permissions
	container.logFile, err = os.Create("out/postgres.log")
	if err != nil {
		tb.Fatalf("error creating output log file for test container: %v", err)
	}
	io.Copy(container.logFile, container.logs)
	if containerErr != nil {
		tb.Fatalf("failed to start postgres test container: %v", containerErr)
	}
	innerPort, err := nat.NewPort("tcp", "5432")
	if err != nil {
		tb.Fatal("error creating port for mapping")
	}
	outerPort, err := container.container.MappedPort(context.Background(), innerPort)
	if err != nil {
		tb.Fatal("error mapping port")
	}

	port := outerPort.Port()
	connectionString = fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", user, password, port, database)
	return
}

func TestMemberRest(t *testing.T) {
	_, connectionString := CreateTestContainer(t)

	err := PerformMigration("../migrations", connectionString)
	if err != nil {
		t.Fatalf("database migration error: %v", err)
	}
	conn, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		t.Fatalf("could not connect to the database")
	}
	defer conn.Close()

	// Clear the database ready for testing
	// _ = conn.QueryRow(context.Background(), "DELETE FROM member;").Scan()

	server := httptest.NewServer(CreateMemberHandler(CreateMemberPgStore(conn), &MemberHandlerConfig{
		DefaultPageSize: 50,
		MaxPageSize:     500,
	}))
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

	t.Run("insert many entities and page through the data", func(t *testing.T) {
		client := TestRestClient{
			t:         t,
			serverUrl: server.URL,
		}

		member := Member{
			FirstName: NewPtr("John"),
			LastName:  NewPtr("Calvin"),
			Notes:     NewPtr("Still writing that very long book"),
		}

		pageSize := 20

		members := make([]Member, 0)

		prevPages := 0
		for client.MakeRequest("GET", fmt.Sprintf("/members?pageSize=%d&page=%d", pageSize, prevPages), nil, &members) != nil &&
			len(members) > 0 {
			prevPages += 1
		}

		ids := make([]uint64, 0)

		for range pageSize + 1 {
			created := Member{}
			_ = client.MakeRequest("POST", "/members", &member, &created)
			if created.Id == nil {
				t.Fatalf("POST returned entity with nil id")
			}

			ids = append(ids, *created.Id)
		}

		pages := 0
		for client.MakeRequest("GET", fmt.Sprintf("/members?pageSize=%d&page=%d", pageSize, pages), nil, &members) != nil &&
			len(members) > 0 {
			pages += 1

			for _, m := range members {
				for iid, id := range ids {
					if id == *m.Id {
						ids = append(ids[:iid], ids[iid+1:]...)
						break
					}
				}
			}
		}

		if len(ids) > 0 {
			t.Errorf("ids %v were created but not found in the paginated result", ids)
		}

		if pages != prevPages+1 {
			t.Errorf("expected to increase page count by 1, but increased it by %d", pages-prevPages)
		}
	})
}
