# FlowState
<sup>Golang stdlib-only</sup>

> [!NOTE]
> Constraint: This project strictly uses the Go Standard Library. No external dependencies are permitted, ensuring a deep dive into core system engineering principles and the Go runtime.

FlowState is a dynamic L7 Reverse Proxy built in Go using only the stdlib.

## Idea

### Design goals

- Build a **stdlib-only Layer 7 reverse proxy** to understand HTTP internals, connection reuse, and request lifecycle.
- Implement **explicit routing, backend selection, and middleware composition** without framework abstractions.
- Make **timeouts, cancellation, health state, and observability** first-class so behavior under failure is understandable.
- Favor **clear control flow and measurable correctness** over feature count.

### Non-goals

- Not a production replacement for Envoy, NGINX, or Traefik.
- No dynamic control plane, distributed config, or service discovery at first.
- No HTTP/3, gRPC-native balancing, or WASM/plugin ecosystem.
- No external metrics/tracing libraries.

### Architecture

- Ingress path: `net/http` server accepts request.
- Router: matches host/path/method to a route entry.
- Middleware chain: request ID, logging, rate limiting, recovery, optional auth later.
- Load balancer: selects a healthy backend for the route.
- Proxy layer: rewrites request, injects forwarding headers, strips hop-by-hop headers, forwards via custom `http.RoundTripper`.
- Health subsystem: active probes + passive failure observation update backend state.
- Metrics/admin endpoints: expose counters, latency buckets, backend health.

Internal flow: `server -> router -> middleware -> lb -> proxy transport -> response filters -> metrics`

### Failure model

- If a backend times out, connection fails, or returns repeated 5xx responses, it may be marked unhealthy and removed from rotation.
- If all backends are unavailable, return `502` or `503` consistently.
- Client cancellation propagates to backend requests through `context.Context`.
- On shutdown, stop accepting new requests and allow in-flight requests to drain for a bounded time.
- No guarantee of zero request loss during process crash or forced termination.

### Benchmark plan

Measure with `testing.B`, `httptest`, and simple load tools.
- Requests/sec through proxy with 1, 10, 100 concurrent clients.
- Latency: p50/p95/p99 under healthy backends.
- Tail latency under one slow backend.
- Cost of middleware stack.
- Effect of connection reuse vs disabled keep-alive.
- Load-balancer distribution fairness across backends.

### Test strategy

- Unit tests for router matching, backend selection, health transitions, middleware behavior.
- `httptest.Server` integration tests for end-to-end proxying.
- Timeout/cancellation tests with slow handlers and canceled contexts.
- Race-detector runs for backend state, counters, and health updates.
- Fuzz tests for route parsing, config parsing, and header rewriting.
- Deterministic failure injection for connection errors and backend flapping.

---

### Core (v0.1)

- [X] **Static config loader**
  - Use: `encoding/json` or `encoding/gob`/custom parser, `os`, `flag`
  - Learn: config modeling, validation, immutable runtime config
- [ ] **HTTP server bootstrap**
  - Use: `net/http`, `context`, `os/signal`
  - Learn: server lifecycle, graceful shutdown, deadline propagation
- [ ] **Basic route matching**
  - Start simple: exact path + prefix match, host/method aware
  - Use: slices/maps, longest-prefix match
  - Learn: request dispatch, matching semantics
- [ ] **Custom reverse proxy path**
  - Use: `net/http`, `net/url`, `io`, custom handler instead of magic wrappers
  - Learn: request cloning, upstream URL rewriting, response copying
- [ ] **Custom transport**
  - Use: `http.Transport`
  - Learn: keep-alives, `MaxIdleConns`, `MaxIdleConnsPerHost`, dial timeout, TLS handshake timeout, response header timeout
- [ ] **Forwarding and hop-by-hop header handling**
  - Learn: `X-Forwarded-For`, `X-Forwarded-Proto`, RFC hop-by-hop header stripping
- [ ] **Weighted round robin**
  - Learn: stateful load balancing, atomic counters vs mutexes, fairness

### Correct (v1.0)

- [ ] **Request-scoped context and timeouts**
  - Use: `context.WithTimeout`
  - Learn: cancellation propagation, bounded resource usage
- [ ] **Structured logging**
  - Use: `log/slog`
  - Learn: request lifecycle logging, correlation IDs
- [ ] **Request ID middleware**
  - Use: headers + context values
  - Learn: cross-cutting concerns, observability basics
- [ ] **Recovery middleware**
  - Use: `defer`, `recover`
  - Learn: panic boundaries in servers
- [ ] **Active health checks**
  - Use: `time.Ticker`, `net/http` or `net.DialTimeout`
  - Learn: health probe design, state transitions
- [ ] **Passive health checks**
  - Learn: consecutive failure counting, decay/reset, backend ejection
- [ ] **Consistent unhealthy-backend policy**
  - Learn: fail-open vs fail-closed, cooldowns
- [ ] **Graceful shutdown and backend draining**
  - Use: `http.Server.Shutdown`
  - Learn: lifecycle management under load
- [ ] **Metrics endpoint**
  - Use: `expvar` or custom text format via `net/http`
  - Learn: counters, gauges, latency buckets

### Performance (v1.1)

- [ ] **Trie router only if route count justifies it**
  - Learn: radix tree / trie, path parameter trade-offs
- [ ] **Power of Two Choices**
  - Learn: queue-aware balancing, herd effect mitigation
- [ ] **Token bucket rate limiter**
  - Use: `time`, `sync`, monotonic refill math
  - Learn: admission control, backpressure
- [ ] **Connection pool tuning benchmarks**
  - Learn: transport tuning under concurrency
- [ ] **Latency histograms / approximate percentiles**
  - Learn: buckets vs exact percentiles, tail latency measurement
- [ ] **Zero-allocation hot-path cleanup where useful**
  - Learn: buffer reuse, avoiding needless header/value copies

### Advanced (v2.0)

- [ ] **Circuit breaker**
  - Learn: closed/open/half-open FSM, rolling failure windows
- [ ] **Hot config reload**
  - Use: atomic config swap, signal-driven reload
  - Learn: read-copy-update style config management
- [ ] **Per-route middleware policies**
  - Learn: policy composition and precedence
- [ ] **Retry policy for idempotent requests only**
  - Learn: retry safety, duplicate side effects
- [ ] **Outlier detection**
  - Learn: passive latency/error-based backend suppression
