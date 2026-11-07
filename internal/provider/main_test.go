package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	kuma "github.com/breml/go-uptime-kuma-client"

	"github.com/breml/terraform-provider-uptimekuma/internal/client"
)

// The account the acceptance tests sign in as. The password satisfies the
// rules an Uptime Kuma 3.x server applies to the first user (at least eight
// characters); `admin1` cannot be created there. The leak check is turned off
// on the container, see kumaContainerEnv.
const (
	username = "admin"
	password = "Kuma-Test-8f4Q2xR9p"
)

// defaultKumaImage is the Uptime Kuma the acceptance tests run against unless
// UPTIME_KUMA_IMAGE names another one, such as louislam/uptime-kuma:nightly3
// for the 3.x line. The CI workflow reads this constant to pre-pull the image.
const defaultKumaImage = "louislam/uptime-kuma:2.5.9"

// kumaImage returns the repository and tag of the Uptime Kuma image to test
// against.
func kumaImage() (repository string, tag string) {
	image := os.Getenv("UPTIME_KUMA_IMAGE")
	if image == "" {
		image = defaultKumaImage
	}

	repository, tag, ok := strings.Cut(image, ":")
	if !ok {
		return image, "latest"
	}

	return repository, tag
}

// kumaContainerEnv is the environment of the test container: a 3.x server
// checks the first user's password against the Have I Been Pwned database,
// which the test password has no business with. 2.x ignores the variable.
func kumaContainerEnv() []string {
	return []string{"UPTIME_KUMA_DISABLE_LEAK_PWD_CHECK=1"}
}

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

		repository, tag := kumaImage()

		// pulls an image, creates a container based on it and runs it
		container, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: repository,
			Tag:        tag,
			Env:        kumaContainerEnv(),
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

		// No container.Expire here, on purpose. dockertest implements it as
		// `go StopContainer(id, seconds)`, an immediate
		// POST /containers/{id}/stop?t=seconds, where `seconds` is only the
		// SIGTERM-to-SIGKILL grace and not a delay before stopping. Nothing
		// about it schedules a stop `seconds` from now, so it was never the
		// safety net it looked like: what it asks for is a `docker stop` right
		// now, which for a suite that needs the container for the whole run is
		// the opposite of what was wanted. (Why it never actually killed a run
		// is unclear - a direct probe against the image showed the container
		// still up and its shutdown handler never entered - so this is about
		// dropping a call that cannot do the job, not about fixing a failure.)
		//
		// The deferred purge below is the real cleanup, and the timeout-minutes
		// on the "Run acceptance tests" step in .github/workflows/test.yml is
		// the outer backstop.

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
		//
		// Together with the pooled provider connection this is all the suite
		// signs in. Keep it that way: a 3.x server allows three sign-ins per
		// ten seconds per address, so per-test logins would get rate limited.
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
