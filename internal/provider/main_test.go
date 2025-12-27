package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	kuma "github.com/breml/go-uptime-kuma-client"

	"github.com/breml/terraform-provider-uptimekuma/internal/client"
)

const (
	username = "admin"
	password = "admin1"
)

var endpoint string //nolint:gochecknoglobals // OK in tests.

func TestMain(m *testing.M) {
	exitCode := runTests(m)
	os.Exit(exitCode)
}

func runTests(m *testing.M) (exitcode int) {
	// We only start the docker based test application, if the TF_ACC env var is
	// set because they're slow.
	if os.Getenv(resource.EnvTfAcc) != "" {
		// uses a sensible default on windows (tcp/http) and linux/osx (socket)
		pool, err := dockertest.NewPool("")
		if err != nil {
			log.Fatalf("Could not construct pool: %v", err)
		}

		// uses pool to try to connect to Docker
		err = pool.Client.Ping()
		if err != nil {
			log.Fatalf("Could not connect to Docker: %v", err)
		}

		// pulls an image, creates a container based on it and runs it
		container, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "louislam/uptime-kuma",
			Tag:        "2",
		}, func(config *docker.HostConfig) {
			// set AutoRemove to true so that stopped container goes away by itself
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		})
		if err != nil {
			log.Fatalf("Could not start resource: %v", err)
		}

		err = container.Expire(480)
		if err != nil {
			log.Fatalf("Could not set expire on container: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		endpoint = fmt.Sprintf("http://localhost:%s", container.GetPort("3001/tcp"))

		var kumaClient *kuma.Client

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		err = pool.Retry(func() error {
			var err error
			kumaClient, err = kuma.New(
				ctx,
				endpoint,
				username,
				password,
				kuma.WithAutosetup(),
				kuma.WithLogLevel(kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL"))),
			)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			log.Printf("Could not connect to uptime kuma: %v", err)
			return 1 // exitcode
		}

		// Close connection again, after we know, the application is running and
		// auto setup has been performed. We don't need the client anymore,
		// Terraform will establish its own connection via the pool.
		err = kumaClient.Disconnect()
		if err != nil {
			log.Printf("Failed to disconnect from uptime kuma: %v", err)
			return 1 // exitcode
		}

		// As of go1.15 testing.M returns the exit code of m.Run(), so it is safe to use defer here
		defer func() {
			// Close the connection pool before purging the container
			err := client.CloseGlobalPool()
			if err != nil {
				log.Printf("Warning: failed to close connection pool: %v", err)
				exitcode = 1
			}

			err = pool.Purge(container)
			if err != nil {
				log.Fatalf("Could not purge resource: %v", err)
				exitcode = 1
			}
		}()
	}

	// The terraform tests create a fresh connection pool, which we close after
	// all tests have been executed.
	defer func() {
		err := client.CloseGlobalPool()
		if err != nil {
			log.Fatalf("Failed to close connection pool after tests: %v", err)
		}
	}()

	return m.Run()
}

func providerConfig() string {
	return fmt.Sprintf(`
provider "uptimekuma" {
  endpoint = %[1]q
  username = %[2]q
  password = %[3]q
}
`, endpoint, username, password)
}
