# internal/client Package

This package provides a client abstraction layer for connecting to Uptime Kuma with built-in retry logic and connection
pooling support.

## Package Purpose

- Manages connections to the Uptime Kuma API
- Provides exponential backoff retry logic for connection reliability
- Implements connection pooling to prevent rate limiting during acceptance tests
- Abstracts client creation from provider configuration

## Files

- [client.go](client.go) - Client creation with retry logic
- [pool.go](pool.go) - Connection pooling implementation
- [client_test.go](client_test.go) - Client tests
- [pool_test.go](pool_test.go) - Pool tests

## Key Constants

### defaultConnectTimeout

```go
const defaultConnectTimeout = 30 * time.Second
```

Applied automatically when no explicit `ConnectTimeout` is configured (or when a negative value is provided). This
prevents the provider from hanging indefinitely when Uptime Kuma is unreachable. The `effectiveTimeout` helper resolves
the configured value or falls back to `defaultConnectTimeout`. The resolved timeout is the **overall** budget for the
whole connection process (across all retry attempts and backoff). Each individual attempt is bounded by
`min(PerAttemptTimeout, remainingBudget)` via `kuma.WithConnectTimeout`. When `PerAttemptTimeout` is zero, each
attempt may use the full remaining budget.

### DefaultOperationTimeout

```go
const DefaultOperationTimeout = 60 * time.Second
```

Applied when no explicit `OperationTimeout` is configured (zero or negative), resolved by
`effectiveOperationTimeout` and passed on as `kuma.WithOperationTimeout`. It bounds the round trip of a **single**
command - the wait for the server's ack and for the update event confirming it. Without it an ack Uptime Kuma never
sends blocks the operation for as long as the caller's context lives, and Terraform gives a provider no deadline of
its own, so the apply would hang indefinitely.

Two things it is easy to get wrong about the scope:

- It is **not** limited to what happens after connecting. The login and setup that `ConnectTimeout` covers are
  commands too, so this budget bounds them as well; while connecting, the effective bound is whichever of the two
  expires first.
- The bound is on the round trip, not on the whole call. A write that has to queue behind other writes to the list it
  broadcasts spends no budget waiting its turn.

Unlike `MaxRetries`, zero means "not configured" and never "no bound". That matters because
`kuma.WithOperationTimeout` reads zero or less as an opt-out that leaves a command bounded only by the caller's
context: passing a raw zero through would restore exactly the unbounded wait this exists to prevent. A caller that
wants a very long bound configures it explicitly. The upstream client defaults to the same minute, so this constant
is about making the provider's default explicit, documented and part of the pool identity rather than about
supplying a bound that would otherwise be missing.

### defaultMaxRetries

```go
const defaultMaxRetries = 3
```

Applied when no explicit `MaxRetries` is configured (or when a negative value is provided). All retry attempts must
complete within the overall `ConnectTimeout` budget, so the default is intentionally small.

## Key Types

### Config

Configuration for Uptime Kuma client creation.

```go
type Config struct {
    Endpoint             string         // Required: Uptime Kuma server URL
    Username             string         // Optional: Login username
    Password             string         // Optional: Login password
    LogLevel             int            // Optional: Socket.IO logging level
    EnableConnectionPool bool           // For acceptance tests, enables pooling
    ConnectTimeout       time.Duration  // Overall timeout budget across all retry attempts (default: 30s)
    PerAttemptTimeout    time.Duration  // Optional per-attempt cap; defaults to remaining ConnectTimeout budget
    OperationTimeout     time.Duration  // Bound on a single operation once connected (default: 60s)
    MaxRetries           int            // Max retry attempts (default: 3)
}
```

**Usage Notes:**

- `Endpoint` is always required
- `Username` and `Password` are both optional or both required (not one without the other)
- `EnableConnectionPool` is enabled during acceptance tests to prevent "login: Too frequently" errors when pooling
- `ConnectTimeout` defaults to `defaultConnectTimeout` (30s) when zero or negative; this value bounds the overall
  connection process across all retry attempts and backoff
- `PerAttemptTimeout` (when greater than zero) caps the time spent on each individual connection attempt; the effective
  per-attempt timeout is `min(PerAttemptTimeout, remainingBudget)`
- `OperationTimeout` defaults to `DefaultOperationTimeout` (60s) when zero or negative; it bounds the round trip of
  each command, including the login and setup performed while connecting, so a command Uptime Kuma never answers
  fails instead of hanging. It is part of the pool's config identity, so a second provider configuration with a
  different value is rejected rather than silently reusing the first connection

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
    Endpoint:             "https://uptime-kuma.example.com",
    Username:             "admin",
    Password:             "password",
    LogLevel:             0,
    EnableConnectionPool: true,           // Always enabled by provider
    ConnectTimeout:       30 * time.Second, // overall budget; 0 uses defaultConnectTimeout
    OperationTimeout:     60 * time.Second, // per-command bound; 0 or negative uses DefaultOperationTimeout
    MaxRetries:           defaultMaxRetries, // 0 means no retries; negative uses defaultMaxRetries
}

client, err := client.New(ctx, config)
if err != nil {
    // Handle error
}
// Pool manages lifecycle - no manual Disconnect()
```

**Retry Logic:**

- Maximum 3 retry attempts (4 total attempts including first try)
- Exponential backoff: base delay 500ms, multiplied by 2^attempt
- Jitter: ±20% randomization (0.8 to 1.2 multiplier)
- Backoff capped to the remaining overall `ConnectTimeout` budget
- Respects context cancellation during backoff

**Backoff Schedule (with defaults):**

- Attempt 1: Immediate
- Attempt 2: ~500ms (400–600ms with jitter)
- Attempt 3: ~1s (800ms–1.2s with jitter)
- Attempt 4: ~2s (1.6–2.4s with jitter)

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

**Note:** The provider calls this from a goroutine watching the Configure RPC context, which is cancelled as soon as
`Configure` returns — long before the client stops being used. The count therefore pairs with `GetOrCreate` calls, not
with live users. `Release` performs no automatic cleanup at `refs=0`; teardown is explicit, via `CloseGlobalPool()`.
Because the release is asynchronous, a caller that closes the pool immediately after its last `Configure` may still
observe a non-zero count.

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

- Delegates to `Pool.CloseIfUnused`, which checks the reference count and closes
  in one critical section so a concurrent `Release` cannot slip between them
- Returns an error if refs != 0 (indicates leak)

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

    // Wrapped in a *providerData, see "Resource Configure Method" below
    pd := &providerData{client: kumaClient, password: data.Password.ValueString()}
    resp.DataSourceData = pd
    resp.ResourceData = pd
}
```

**Critical Context Detail:**

- Uses `context.Background()` not Terraform's context
- Terraform cancels its context after Configure() completes
- Socket.IO connection must outlive the Configure method
- A goroutine watches the Configure RPC context and calls `Pool.Release` when it
  is cancelled. That happens as soon as Configure returns, so the reference
  count tracks Configure calls rather than live users; the connection itself
  lives until `CloseGlobalPool`

### Resource Configure Method

Resources receive the client via `ProviderData`, which carries a `*providerData`, not
the client itself. `configureClient` unwraps it, so no resource repeats the type
assertion:

```go
func (r *MonitorHTTPResource) Configure(
    _ context.Context,
    req resource.ConfigureRequest,
    resp *resource.ConfigureResponse,
) {
    r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
```

`configureClient` ([../provider/provider_data.go](../provider/provider_data.go)) returns
nil when `ProviderData` is nil, which is how the framework calls Configure before the
provider itself is configured, and raises a diagnostic on any other type.

## Testing Considerations

### Acceptance Test Setup

In [../provider/main_test.go](../provider/main_test.go):

```go
func runTests(m *testing.M) (exitcode int) {
    // ... Docker setup

    // Purge the container and close both connections however we leave.
    defer func() {
        outOfBandClient.Disconnect()
        client.CloseGlobalPool()
        pool.Purge(container)
    }()

    // The first connection runs autosetup, which creates the admin user. It is
    // kept as the out-of-band client for the disappears tests instead of being
    // closed, because creating another one would hit Uptime Kuma's login rate
    // limit.
    pool.Retry(func() error {
        var err error
        outOfBandClient, err = kuma.New(
            context.Background(), endpoint, username, password, kuma.WithAutosetup(),
        )

        return err
    })

    // Run tests. The provider opens the pooled connection itself, from
    // Configure, and every test shares it.
    return m.Run()
}
```

Pooling is not switched on by the tests: the provider always passes
`EnableConnectionPool: true` to `client.New`, so a single connection is shared no
matter how many provider instances Terraform creates.

**Global Variables** (used by all tests):

```go
var (
    endpoint        string       // e.g., "http://localhost:32768"
    outOfBandClient *kuma.Client // deletes resources behind Terraform's back
)

const (
    username = "admin"
    password = "admin1"  // throwaway credentials for the test container
)
```

**Why Pooling?**

- Uptime Kuma rate limits login attempts
- Without pooling, parallel acceptance tests trigger "login: Too frequently" errors
- Single shared connection prevents rate limit issues
- Reference counting helps detect leaks

### Pool Lifecycle

1. **Setup**: Initial client created for autosetup, then kept as the out-of-band client
2. **Tests Run**: The provider's `Configure` passes `EnableConnectionPool: true`
3. **First Configure**: Creates the pooled connection via `GetOrCreate()`
4. **Subsequent Configure calls**: Reuse the existing pooled connection, refs incremented
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
- `"failed after 4 attempts: ..."` - Connection retry exhaustion (with default `max_retries=3`)
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
