package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/carsonalh/churchmanagerbackend/server/controller"
	"github.com/carsonalh/churchmanagerbackend/server/domain"
	"github.com/carsonalh/churchmanagerbackend/server/server"
	"github.com/carsonalh/churchmanagerbackend/server/util"
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

func TestMemberRest(t *testing.T) {
	if TestConnectionString == nil {
		t.Fatal("TestConnectionString is nil; cannot proceed")
	}
	connectionString := *TestConnectionString
	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		t.Fatalf("could not connect to the database")
	}
	defer pool.Close()

	server := httptest.NewServer(server.CreateServer(pool, server.ServerConfig{
		Members: controller.MemberControllerConfig{
			DefaultPageSize: 50,
			MaxPageSize:     500,
		},
	}))
	defer server.Close()

	t.Run("POST and GET again", func(t *testing.T) {
		client := TestRestClient{
			t:         t,
			serverUrl: server.URL,
		}

		requestBody := domain.MemberUpdateDTO{
			FirstName:    util.NewPtr("Thomas"),
			LastName:     util.NewPtr("More"),
			EmailAddress: util.NewPtr("thomas.more.1478@gmail.com"),
			Notes:        "Not to be put in the same Bible study as Luther",
		}

		firstResponse := domain.MemberResponseDTO{}

		response := client.MakeRequest("POST", "/members", &requestBody, &firstResponse)
		if response.StatusCode != http.StatusCreated {
			t.Errorf("expected status 201 Created, but got %s", response.Status)
		}
		location, err := response.Location()
		if err != nil {
			t.Fatalf("could not read Location header from response: %v", err)
		}

		finalResponse := domain.MemberResponseDTO{}
		_ = client.MakeRequest("GET", location.Path, nil, &finalResponse)

		derefOrNil := func(x *string) any {
			if x == nil {
				return nil
			}
			return *x
		}

		if derefOrNil(requestBody.FirstName) != derefOrNil(finalResponse.FirstName) ||
			derefOrNil(requestBody.LastName) != derefOrNil(finalResponse.LastName) ||
			derefOrNil(requestBody.EmailAddress) != derefOrNil(finalResponse.EmailAddress) {
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

		member := domain.MemberUpdateDTO{
			FirstName:    util.NewPtr("Martin"),
			LastName:     util.NewPtr("Luther"),
			EmailAddress: util.NewPtr("martin_luther@live.co.de"),
			PhoneNumber:  util.NewPtr("0428374598"),
		}

		var createdMember domain.MemberResponseDTO

		response := client.MakeRequest("POST", "/members", &member, &createdMember)
		if response.StatusCode != http.StatusCreated {
			t.Errorf("expected POST /members response to be 201 Created but was %s", response.Status)
		}
		createdId := createdMember.Id

		members := make([]domain.MemberResponseDTO, 0)

		response = client.MakeRequest("GET", "/members", nil, &members)
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected GET /members response to be 200 OK but was %s", response.Status)
		}

		recordFound := false
		for _, m := range members {
			if m.Id == createdId {
				if recordFound {
					t.Error("found another record with the id being searched for")
				}
				recordFound = true
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

		member := domain.MemberUpdateDTO{
			FirstName: util.NewPtr("Carson"),
			LastName:  util.NewPtr("Holloway"),
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

		member := domain.MemberUpdateDTO{
			FirstName: util.NewPtr("John"),
			LastName:  util.NewPtr("Calvin"),
			Notes:     "Still writing that very long book",
		}

		response := client.MakeRequest("POST", "/members", &member, &member)
		location, err := response.Location()
		if err != nil {
			t.Fatalf("could not read Location header from response: %v", err)
		}

		member.Notes = "Now re-writing the same book in French"

		response = client.MakeRequest("PUT", location.Path, &member, nil)
		if response.StatusCode != http.StatusOK {
			t.Errorf("PUT /members/{id} : expected status 200 OK but got %s", response.Status)
		}

		response = client.MakeRequest("GET", location.Path, nil, &member)
		if response.StatusCode != http.StatusOK {
			t.Errorf("GET /members/{id} : expected status 200 OK but got %s", response.Status)
		}

		if member.Notes != "Now re-writing the same book in French" {
			t.Error("updated data did not correctly persist accross calls")
		}
	})

	t.Run("POST, DELETE and DELETE returns a 404", func(t *testing.T) {
		client := TestRestClient{
			t:         t,
			serverUrl: server.URL,
		}

		member := domain.MemberRow{
			FirstName: util.NewPtr("John"),
			LastName:  util.NewPtr("Calvin"),
			Notes:     "Still writing that very long book",
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

		member := domain.MemberUpdateDTO{
			FirstName: util.NewPtr("John"),
			LastName:  util.NewPtr("Calvin"),
			Notes:     "Still writing that very long book",
		}

		pageSize := 20

		members := make([]domain.MemberResponseDTO, 0)

		prevPages := 0
		for client.MakeRequest("GET", fmt.Sprintf("/members?pageSize=%d&page=%d", pageSize, prevPages), nil, &members) != nil &&
			len(members) > 0 {
			prevPages += 1
		}

		ids := make([]uint64, 0)

		for range pageSize + 1 {
			created := domain.MemberResponseDTO{}
			_ = client.MakeRequest("POST", "/members", &member, &created)
			ids = append(ids, created.Id)
		}

		pages := 0
		for client.MakeRequest("GET", fmt.Sprintf("/members?pageSize=%d&page=%d", pageSize, pages), nil, &members) != nil &&
			len(members) > 0 {
			pages += 1

			for _, m := range members {
				for iid, id := range ids {
					if id == m.Id {
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
