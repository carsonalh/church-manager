package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/carsonalh/churchmanagerbackend/server/migration"
	"github.com/carsonalh/churchmanagerbackend/server/util"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var TestConnectionString *string = nil

func TestMain(m *testing.M) {
	container, err := CreateTestContainer()
	if err != nil {
		panic(err)
	}

	err = migration.PerformMigration("../../migrations", container.connectionString)
	if err != nil {
		panic(fmt.Errorf("database migration error: %v", err))
	}

	TestConnectionString = util.NewPtr(container.connectionString)

	code := m.Run()

	err = container.logs.Close()
	if err != nil {
		fmt.Printf("failure to close logs from container: %v\n", err)
		code = 1
	}

	err = container.container.Stop(context.Background(), nil)
	if err != nil {
		fmt.Printf("failure to stop container: %v\n", err)
		code = 1
	}

	os.Exit(code)
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
	container        testcontainers.Container
	logs             io.ReadCloser
	logFile          *os.File
	connectionString string
}

func CreateTestContainer() (*TestPostgresContainer, error) {
	user := "postgres"
	password := "admin"
	database := "churchmanager"

	container := new(TestPostgresContainer)

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
	container.logs, err = container.container.Logs(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting logs from test container: %v", err)
	}
	_ = os.Mkdir("out", 0o755) // rwxr-xr-x permissions
	container.logFile, err = os.Create("out/postgres.log")
	if err != nil {
		return nil, fmt.Errorf("error creating output log file for test container: %v", err)
	}
	io.Copy(container.logFile, container.logs)
	if containerErr != nil {
		return nil, fmt.Errorf("failed to start postgres test container: %v", containerErr)
	}
	innerPort, err := nat.NewPort("tcp", "5432")
	if err != nil {
		return nil, fmt.Errorf("error creating port for mapping")
	}
	outerPort, err := container.container.MappedPort(context.Background(), innerPort)
	if err != nil {
		return nil, fmt.Errorf("error mapping port")
	}

	port := outerPort.Port()
	container.connectionString = fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", user, password, port, database)
	return container, nil
}
