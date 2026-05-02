# Config v0.1

## Package responsibility

The `internal/config` package is responsible for:

- reading configuration contents from disk
- decoding JSON into typed Go structs
- validating configuration fields and structure
- returning a typed configuration object to the caller

## Model

[Models.md](./Models.md) is defining the first expectations we have for the configuration.

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

The configuration is quite basic for now, but represents most of the features we might want to use for a very simple reverse proxy.

We chose JSON cause it is easily workable with when using only stdlibs. We all know that JSON is a mess to maintain in time and not the best for humans, however it is the simplest approach for v0.1.
YAML is broadly adopted nowadays for easy configurations but it is harder to support with stdlib, this can be an improvement for later.

## Static config loader

First we need to choose what encoding we will use.

<details>
<summary>[encoding/json](https://pkg.go.dev/encoding/json@go1.26.2)</summary>

`encoding/json` is the stable standard library JSON package. It follows [RFC 7159](https://rfc-editor.org/rfc/rfc7159.html), is widely used, and prioritizes backwards compatibility.

It preserves some legacy behaviors for compatibility, including edge cases that can be surprising, but it remains the conservative and practical choice for configuration loading.

</details>

<details>
<summary> :x: [encoding/json/v2](https://pkg.go.dev/encoding/json/v2@go1.26.2)</summary>

`encoding/json/v2` is the newer JSON API intended to address limitations and legacy behaviors in `encoding/json`. It follows [RFC 8259](https://rfc-editor.org/rfc/rfc8259.html) and offers a more modern and explicit design.

For this project, however, it is not the right choice. The goal is not to perform an in-depth study of JSON semantics or to explore the newer API surface. We want the most stable, conventional, and low-friction option for static configuration parsing.

As of now, it is not the conservative default choice for this project.

</details>

Then we will make use of the [os](https://pkg.go.dev/os@go1.26.2) lib: to open and read the configuration from disk.
