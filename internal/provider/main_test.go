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
)

const (
	username = "admin"
	password = "admin1"
)

var endpoint string

func TestMain(m *testing.M) {
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
		resource, err := pool.RunWithOptions(&dockertest.RunOptions{
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

		err = resource.Expire(240)
		if err != nil {
			log.Fatalf("Could not set expire on container: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		endpoint = fmt.Sprintf("http://localhost:%s", resource.GetPort("3001/tcp"))

		var client *kuma.Client

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		if err := pool.Retry(func() error {
			var err error
			client, err = kuma.New(
				ctx,
				endpoint, username, password,
				kuma.WithAutosetup(),
				kuma.WithLogLevel(kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL"))),
			)
			if err != nil {
				return err
			}

			return nil
		}); err != nil {
			log.Fatalf("Could not connect to uptime kuma: %v", err)
		}

		settings, err := client.GetSettings(ctx)
		if err != nil {
			log.Fatalf("Failed to get settings: %v", err)
		}

		settings["disableAuth"] = true

		err = client.SetSettings(ctx, settings, password)
		if err != nil {
			log.Fatalf("Failed to set settings: %v", err)
		}

		// Close connection again, after we know, the application is running and
		// auto setup has been performed. We don't need the client anymore,
		// Terraform will establish its own connection.
		err = client.Disconnect()
		if err != nil {
			log.Fatalf("Failed to connect to uptime kuma: %v", err)
		}

		// As of go1.15 testing.M returns the exit code of m.Run(), so it is safe to use defer here
		defer func() {
			if err := pool.Purge(resource); err != nil {
				log.Fatalf("Could not purge resource: %v", err)
			}
		}()
	}

	m.Run()
}

func providerConfig() string {
	return fmt.Sprintf(`
provider "uptimekuma" {
  endpoint = %[1]q
}
`, endpoint, username, password)
}
