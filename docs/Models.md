# FlowState Models

This document defines the initial runtime and configuration models for FlowState milestone 0. The goal is to keep the system small, explicit, and easy to evolve.

## Design intent

FlowState starts as a static-config Layer 7 reverse proxy with:

- host-based routing
- path-prefix routing
- optional HTTP method filtering
- predefined middleware chains
- weighted round-robin backend selection

The models below define the core behavior before implementation.

---

## 1. Configuration model

Configuration is loaded from a static JSON file using the Go standard library only.

### Top-level shape

```json
{
  "server": {
    "listen_addr": ":8080",
    "read_timeout_ms": 5000,
    "write_timeout_ms": 10000,
    "idle_timeout_ms": 60000
  },
  "routes": [
    {
      "host": "api.local",
      "path_prefix": "/users",
      "methods": ["GET", "POST"],
      "middlewares": ["request_id", "logging"],
      "backends": [
        { "url": "http://127.0.0.1:9001", "weight": 2 },
        { "url": "http://127.0.0.1:9002", "weight": 1 }
      ]
    }
  ]
}
```

### Server config

The server section configures the public HTTP listener.

Fields:
- `listen_addr`: address passed to the HTTP server
- `read_timeout_ms`: max duration to read request headers/body
- `write_timeout_ms`: max duration to write the response
- `idle_timeout_ms`: keep-alive idle timeout

### Route config

Each route defines how incoming requests are matched and where they are forwarded.

Fields:
- `host`: incoming `Host` header to match
- `path_prefix`: path prefix to match
- `methods`: optional allowed HTTP methods
- `middlewares`: predefined middleware names applied in order
- `backends`: upstream targets for this route

### Backend config

Each backend represents one upstream server.

Fields:
- `url`: upstream base URL
- `weight`: relative load-balancing weight

Notes:
- weight must be positive
- runtime health state is not part of config

---

## 2. Route model

A route is a matching rule plus a forwarding policy.

Initial route identity:
- host
- path prefix
- optional method set
- middleware list
- backend pool

### Matching rules

A request matches a route when:
- request host equals route host
- request path starts with route path prefix
- request method is in route methods if methods are provided

### Route precedence

If multiple routes match:
- the route with the **longest path prefix** wins

This keeps routing deterministic and simple.

### Initial scope

Supported:
- exact host match
- path prefix match
- optional method filtering

Not supported initially:
- path parameters
- regex routes
- wildcard hosts
- query-based routing
- header-based routing

---

## 3. Backend model

A backend has both config state and runtime state.

### Static backend fields

- `url`
- `weight`

### Runtime backend fields

Runtime-only state should be tracked separately from config.

Suggested fields:
- healthy/unhealthy
- consecutive failure count
- last failure time
- last health-check time
- in-flight request count, later if needed

### Backend selection

Initial strategy:
- weighted round robin

Later candidates:
- power of two choices
- least loaded
- latency-aware selection

---

## 4. Timeout model

Timeouts must be explicit because proxy correctness depends on bounded waiting.

### Inbound server timeouts

Configured on the public HTTP server:
- read timeout
- write timeout
- idle timeout

### Upstream transport timeouts

Configured on the custom `http.Transport`:
- dial timeout
- TLS handshake timeout
- response header timeout
- idle connection timeout

### Optional request timeout

A request may also receive a per-request deadline through `context.Context`.

Initial rule:
- client cancellation must propagate to the upstream request

---

## 5. Error model

FlowState should respond consistently under failure.

### Error mapping

- no matching route -> `404 Not Found`
- route matched but no healthy backend -> `503 Service Unavailable`
- upstream connection failure -> `502 Bad Gateway`
- upstream timeout -> `504 Gateway Timeout`
- middleware rejection -> middleware-defined status, e.g. `429 Too Many Requests`
- internal panic -> `500 Internal Server Error`

### Principles

- client-facing errors should be deterministic
- internal error details should be logged, not exposed
- backend failures should update runtime health state when appropriate

---

## 6. Middleware model

Middleware is configured by predefined names in the route config.

Initial examples:
- `request_id`
- `logging`
- `recovery`

### Rules

- middleware order is explicit in config
- middleware wraps the route handler in declared order
- middleware should remain stateless unless state is required by design, such as rate limiting

Not supported initially:
- arbitrary user-defined plugins
- dynamic middleware loading

---

## 7. Observability model

Initial observability should be minimal but useful.

Track first:
- total request count
- request count by status code
- request duration
- backend health state

Possible implementation:
- in-memory counters
- `expvar`
- custom `/metrics` text endpoint
