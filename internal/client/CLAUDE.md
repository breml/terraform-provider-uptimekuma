# internal/client Package

This package provides a client abstraction layer for connecting to Uptime Kuma with built-in retry logic and connection
pooling support.

## Package Purpose

- Manages connections to the Uptime Kuma API
- Provides exponential backoff retry logic for connection reliability
- Implements connection pooling to prevent rate limiting during acceptance tests
- Abstracts client creation from provider configuration

## Files

- [client.go](client.go) - Client creation with retry logic (79 lines)
- [pool.go](pool.go) - Connection pooling implementation (156 lines)
- [client_test.go](client_test.go) - Client tests
- [pool_test.go](pool_test.go) - Pool tests

## Key Types

### Config

Configuration for Uptime Kuma client creation.

```go
type Config struct {
    Endpoint             string  // Required: Uptime Kuma server URL
    Username             string  // Optional: Login username
    Password             string  // Optional: Login password
    LogLevel             int     // Optional: Socket.IO logging level
    EnableConnectionPool bool    // For acceptance tests, enables pooling
}
```

**Usage Notes:**

- `Endpoint` is always required
- `Username` and `Password` are both optional or both required (not one without the other)
- `EnableConnectionPool` is enabled during acceptance tests to prevent "login: Too frequently" errors when pooling

### Pool

Manages a shared Socket.IO connection for test scenarios.

```go
type Pool struct {
    mu     sync.Mutex      // Protects all fields
    client *kuma.Client    // Shared client instance
    config *Config         // Config used to create client
    refs   int             // Reference count for debugging
}
```

## Client Creation Patterns

### Pooled Connection (Provider Use)

Used by the provider in both production and testing:

```go
config := &Config{
    Endpoint: "https://uptime-kuma.example.com",
    Username: "admin",
    Password: "password",
    LogLevel: 0,
    EnableConnectionPool: true,  // Always enabled by provider
}

client, err := client.New(ctx, config)
if err != nil {
    // Handle error
}
// Pool manages lifecycle - no manual Disconnect()
```

**Retry Logic:**

- Maximum 5 retry attempts (6 total attempts including first try)
- Exponential backoff: base delay 5 seconds, multiplied by 2^attempt
- Jitter: ±20% randomization (0.8 to 1.2 multiplier)
- Maximum backoff capped at 30 seconds
- Respects context cancellation during backoff

**Backoff Schedule:**

- Attempt 1: Immediate
- Attempt 2: ~5s (4-6s with jitter)
- Attempt 3: ~10s (8-12s with jitter)
- Attempt 4: ~20s (16-24s with jitter)
- Attempt 5: ~30s (capped)
- Attempt 6: ~30s (capped)

### Acceptance Tests

During acceptance tests, the same pooled configuration is used:

```go
config := &Config{
    Endpoint: "http://localhost:3001",
    Username: "admin",
    Password: "admin",  // Same as main_test.go
    EnableConnectionPool: true,
}

client, err := client.New(ctx, config)
if err != nil {
    // Handle error
}
// Pool manages lifecycle - connection shared across all tests
```

**Pool Behavior:**

- First call creates new connection using same retry logic
- Subsequent calls return existing connection (if config matches)
- Config validation prevents credential confusion
- Reference counting tracks usage (for debugging)
- Connection persists until explicit pool closure

## Connection Pool Implementation

### Singleton Pattern

The pool uses `sync.Once` to ensure only one global pool instance exists:

```go
func GetGlobalPool() *Pool
```

**Global State:**

- `globalPool`: The singleton pool instance
- `globalPoolOnce`: Ensures single initialization
- `globalPoolMu`: Protects pool and once during reset

### Pool Operations

#### GetOrCreate

Returns existing client or creates new one:

```go
client, err := pool.GetOrCreate(ctx, config)
```

**Validation:**

- If pool has existing client, validates config matches (endpoint, username, password)
- LogLevel not validated (first connection's level is used)
- Increments reference count on success
- Returns error on config mismatch to prevent credential confusion

#### Release

Decrements reference count (for debugging):

```go
pool.Release()
```

**Note:** Used by the provider during shutdown (via `client.GetGlobalPool().Release()`) to decrement the global pool's
reference count before `CloseGlobalPool()` is called. `Release` itself does not perform automatic cleanup at `refs=0`;
global pool teardown is handled explicitly by `CloseGlobalPool()`, which is expected to run after all matching
`Acquire`/`Release` calls for a test sequence.

#### Close

Forcefully closes pooled connection and resets pool state:

```go
err := pool.Close()
```

**Behavior:**

- Calls `client.Disconnect()` on underlying Socket.IO connection
- Resets all pool fields to nil/zero
- Should only be called during test cleanup

### Global Pool Management

Convenience functions for test lifecycle:

```go
// Close the global pool (called in TestMain cleanup)
err := CloseGlobalPool()

// Reset for test isolation (used in individual tests if needed)
ResetGlobalPool()
```

**CloseGlobalPool Validation:**

- Checks reference count is 0 before closing
- Returns error if refs != 0 (indicates leak)

## Integration with Provider

### Provider Configure Method

In [../provider/provider.go](../provider/provider.go), the provider creates a client:

```go
func (p *UptimeKumaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    // ... read config from req.Config

    // Use context.Background() not ctx (Terraform's context cancels too early)
    kumaClient, err := client.New(context.Background(), &client.Config{
        Endpoint:             data.Endpoint.ValueString(),
        Username:             data.Username.ValueString(),
        Password:             data.Password.ValueString(),
        EnableConnectionPool: true,  // Always enabled
        LogLevel:             kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
    })

    // Make client available to both data sources and resources
    resp.DataSourceData = kumaClient
    resp.ResourceData = kumaClient
}
```

**Critical Context Detail:**

- Uses `context.Background()` not Terraform's context
- Terraform cancels its context after Configure() completes
- Socket.IO connection must outlive the Configure method
- Connection lifetime managed by goroutine watching for actual cleanup signal

### Resource Configure Method

Resources receive the client via `ProviderData`:

```go
func (r *MonitorHTTPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*kuma.Client)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *kuma.Client, got: %T", req.ProviderData),
        )
        return
    }

    r.client = client
}
```

## Testing Considerations

### Acceptance Test Setup

In [../provider/main_test.go](../provider/main_test.go):

```go
func TestMain(m *testing.M) {
    // ... Docker setup

    // Create initial client for autosetup
    kumaClient, err := kuma.New(context.Background(), endpoint, username, password)

    // Close initial connection
    kumaClient.Disconnect()

    // Enable pooling for actual tests
    enableConnectionPool = true

    // Run tests (all will share pooled connection)
    code := m.Run()

    // Cleanup
    CloseGlobalPool()
    pool.Purge(resource)
}
```

**Global Variables** (used by all tests):

```go
var (
    endpoint string  // e.g., "http://localhost:32768"
)

const (
    username = "admin"
    password = "password123"
)
```

**Why Pooling?**

- Uptime Kuma rate limits login attempts
- Without pooling, parallel acceptance tests trigger "login: Too frequently" errors
- Single shared connection prevents rate limit issues
- Reference counting helps detect leaks

### Pool Lifecycle

1. **Setup**: Initial client created for database autosetup, then closed
2. **Tests Run**: `enableConnectionPool = true` set globally
3. **First Test**: Creates pooled connection via `GetOrCreate()`
4. **Subsequent Tests**: Reuse existing pooled connection
5. **Cleanup**: `CloseGlobalPool()` called, validates refs=0, disconnects

### Test Isolation

If individual tests need pool isolation:

```go
func TestSomething(t *testing.T) {
    // Reset pool state before test
    ResetGlobalPool()
    defer CloseGlobalPool()

    // Test code
}
```

## Error Handling Patterns

### Client Creation Errors

```go
client, err := client.New(ctx, config)
if err != nil {
    return fmt.Errorf("create uptime kuma client: %w", err)
}
```

**Common Errors:**

- `"endpoint is required"` - Config validation failure
- `"failed after 6 attempts: ..."` - Connection retry exhaustion
- `"connection cancelled: ..."` - Context cancellation during retry
- `"pool config mismatch: ..."` - Credential confusion prevention

### Pool Errors

```go
err := CloseGlobalPool()
if err != nil {
    // "failed to close global pool, expected 0 refs, got: N"
    // Indicates potential resource leak
}
```

## Dependencies

- `github.com/breml/go-uptime-kuma-client` - Uptime Kuma API client
  - Provides `kuma.Client` type and `kuma.New()` constructor
  - Socket.IO-based real-time connection
  - `WithLogLevel()` option for debugging

## Design Decisions

### Why context.Background()?

Terraform's context is cancelled immediately after Configure() completes. Socket.IO connections must outlive this method
to be usable by resources. A separate goroutine monitors for actual cleanup signals.

### Why Singleton Pool?

Acceptance tests run in a single process with multiple provider instances. A singleton ensures all providers share one
connection, preventing rate limiting.

### Why Reference Counting?

Primarily for debugging and detecting leaks. Currently not used for automatic cleanup, but provides visibility into pool
usage patterns.

### Why Not Close Connection Between Tests?

Creating/destroying connections is slow and triggers rate limiting. Keeping one connection alive across all tests is
faster and more reliable.

### Why Config Validation?

Prevents credential confusion where different tests might try to use the pool with different endpoints or credentials.
Fails fast with clear error message.

## Future Enhancements

Potential improvements (not currently implemented):

1. **Automatic Cleanup**: Use reference counting to automatically close connection when refs reach 0
2. **Multiple Pools**: Support multiple pools for different endpoints
3. **Connection Health Checks**: Verify connection is still alive before returning from pool
4. **Metrics**: Track connection attempts, failures, pool hits/misses
5. **Graceful Degradation**: Fall back to direct connection if pool fails
