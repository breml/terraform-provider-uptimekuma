package provider

import (
	"context"
	"os"
	"testing"
	"time"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccConnectionStability verifies that repeated sequential connections
// to Uptime Kuma complete reliably. This exercises the Socket.IO transport
// upgrade path (polling → websocket) multiple times, which amplifies the
// probability of hitting the race conditions fixed in the breml/go.socket.io
// fork (see issue #283). Combined with Go's -race detector, this test
// catches unsynchronized access in the socket.io transport layer.
func TestAccConnectionStability(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skip("TF_ACC not set — skipping acceptance test")
	}

	if endpoint == "" {
		t.Fatal("endpoint not set — TestMain did not initialize the Docker container")
	}

	const iterations = 10
	const perConnTimeout = 30 * time.Second

	for i := range iterations {
		// Wrap each iteration in a context.WithTimeout so that a regression
		// (event dispatch hang after transport establishment) fails fast
		// instead of blocking until the global go test -timeout fires.
		ctx, cancel := context.WithTimeout(t.Context(), perConnTimeout)

		kumaClient, err := kuma.New(
			ctx,
			endpoint,
			username,
			password,
			kuma.WithConnectTimeout(perConnTimeout),
			kuma.WithLogLevel(kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL"))),
		)
		if err != nil {
			cancel()
			t.Fatalf("connection %d/%d failed: %v", i+1, iterations, err)
		}

		err = kumaClient.Disconnect()

		cancel()

		if err != nil {
			t.Fatalf("disconnect %d/%d failed: %v", i+1, iterations, err)
		}
	}
}
