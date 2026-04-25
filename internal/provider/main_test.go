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

// outOfBandClient is a dedicated kuma client for out-of-band operations in
// acceptance tests (e.g. deleting resources externally in disappears tests).
// It is separate from the provider's pooled connection to genuinely simulate
// external side effects, and is created once in TestMain to avoid Uptime
// Kuma's login rate limiting.
var outOfBandClient *kuma.Client //nolint:gochecknoglobals // OK in tests.

func TestMain(m *testing.M) {
	//nolint:revive // Exit is needed to propagate exitcode set by deferred cleanup.
	os.Exit(runTests(m))
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
			Tag:        "2.2.0",
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

		err = container.Expire(1200)
		if err != nil {
			log.Fatalf("Could not set expire on container: %v", err)
		}

		// Register container cleanup immediately so the container is always
		// purged, even if the connection retry below fails.
		defer func() {
			// Close the out-of-band client (may be nil if connection failed).
			if outOfBandClient != nil {
				disconnectErr := outOfBandClient.Disconnect()
				if disconnectErr != nil {
					log.Printf("Warning: failed to disconnect out-of-band client: %v", disconnectErr)
					exitcode = 1
				}
			}

			// Close the connection pool before purging the container.
			closeErr := client.CloseGlobalPool()
			if closeErr != nil {
				log.Printf("Warning: failed to close connection pool: %v", closeErr)
				exitcode = 1
			}

			purgeErr := pool.Purge(container)
			if purgeErr != nil {
				log.Printf("Warning: could not purge resource: %v", purgeErr)
				exitcode = 1
			}
		}()

		endpoint = fmt.Sprintf("http://localhost:%s", container.GetPort("3001/tcp"))

		// exponential backoff-retry, because the application in the container
		// might not be ready to accept connections yet. This first connection
		// performs autosetup (creating the admin user) and is kept as the
		// out-of-band client for disappears tests.
		err = pool.Retry(func() error {
			var retryErr error
			outOfBandClient, retryErr = kuma.New(
				context.Background(),
				endpoint,
				username,
				password,
				kuma.WithAutosetup(),
				kuma.WithLogLevel(kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL"))),
				kuma.WithConnectTimeout(30*time.Second),
			)

			return retryErr
		})
		if err != nil {
			log.Printf("Could not connect to uptime kuma: %v", err)
			return 1 // exitcode
		}
	}

	// The terraform tests create a fresh connection pool, which we close after
	// all tests have been executed. Use log.Printf + exitcode instead of
	// log.Fatalf to avoid os.Exit skipping earlier deferred cleanups.
	defer func() {
		err := client.CloseGlobalPool()
		if err != nil {
			log.Printf("Failed to close connection pool after tests: %v", err)
			exitcode = 1
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
