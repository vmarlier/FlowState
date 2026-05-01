# FlowState
<sup>Golang</sup> <sub>stdlib only</sub>

FlowState is a dynamic Layer 7 Reverse Proxy and API Gateway developed exclusively with the Go standard library. It acts as a resilient entry point for microservices, providing intelligent request routing, weighted load balancing, and automated backend health monitoring. With a focus on observability and traffic control, FlowState features a modular middleware pipeline that handles rate limiting, circuit breaking, and request transformation, allowing for sophisticated traffic management at the edge of the network.

> [!NOTE]
> Constraint: This project strictly uses the Go Standard Library. No external dependencies are permitted, ensuring a deep dive into core system engineering principles and the Go runtime.

## Todo

- [ ] Custom HTTP Transport: Implement a custom `http.RoundTripper`. Do not just use the default; configure connection pooling (MaxIdleConns), timeouts, and TLS handshakes.
- [ ] The Router (Trie-based): Build a high-performance router. Standard `http.ServeMux` is too simple for a gateway.
- [ ] Load Balancing (Stateful): Implement Weighted Round Robin and P2C (Power of Two Choices) to avoid the "herd effect."
- [ ] Middleware Chain: Build a recursive or slice-based middleware wrapper.
  - Rate Limiter: Implement a Token Bucket or Leaky Bucket algorithm.
  - Circuit Breaker: Implement a state machine (Closed, Open, Half-Open) to prevent cascading failures.
- [ ] Active & Passive Health Checks: * Active: Periodically dial backends.Passive: If a backend returns 5xx errors $N$ times, pull it from the rotation.
- [ ] Request Transformation: Logic to strip prefixes, inject `X-Forwarded-For` headers, and handle Hop-by-hop headers.
- [ ] Observability: Create a `/metrics` endpoint that exports basic stats (request count, latency p99, error rates) in a format like Prometheus.

### Knowledge kit

- <ins>Go Packages:</ins> `net/http`, `net/http/httputil`, `context`, `crypto/tls`, `net/url`.
- <ins>Algorithms:</ins> Trie (Prefix Tree), Token Bucket, Power of Two Choices (P2C).
- <ins>Patterns:</ins> Decorator Pattern, Strategy Pattern (for LB), Circuit Breaker.
- <ins>Keywords:</ins> L7 vs L4 Proxy, Connection Pooling, TLS Termination, Keep-Alive, Backpressure, Idempotency.
