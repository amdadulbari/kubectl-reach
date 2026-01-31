# kubectl-reach

A `kubectl` plugin that tests network connectivity **from** an existing Pod **to** a target (IP, DNS, or Service) by injecting an ephemeral debug container into the source Pod. Useful for verifying if NetworkPolicies or Service Meshes are blocking traffic.

## Requirements

- **Kubernetes 1.25+** (EphemeralContainers is stable).
- No shell dependencies; pure Go. Works on Linux, macOS, and Windows.

## Usage

The plugin runs **from** a given Pod (by name) and checks whether it can reach a target `host:port`. The target can be an IP, a DNS name, or a Kubernetes Service. It uses an ephemeral debug container in the source Pod, so the check runs in the same network context (NetworkPolicies, Service Mesh, node) as your workload.

```bash
kubectl reach <source-pod-name> --to <target:port> [--namespace <ns>] [options]
```

### Examples

```bash
# Reach a public host (HTTPS)
kubectl reach myapp-abc123 --to google.com:443

# Reach an internal IP and port (default namespace)
kubectl reach myapp-abc123 --to 10.0.0.5:8080

# Specify namespace
kubectl reach myapp-abc123 --to myservice:80 -n myns

# Use a custom debug image and longer timeout
kubectl reach myapp-abc123 --to myservice:80 --image busybox --timeout 10
```

On success you see an "open" result; on failure you see the error (e.g. connection refused, timed out). Use your usual kubectl context/namespace (`-n`, `--context`, `--kubeconfig`) to target the right cluster and namespace.

### Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--to` | Yes | - | Target as `host:port` (e.g. `google.com:443`, `10.0.0.5:8080`) |
| `--image` | No | `busybox` | Debug container image |
| `--timeout` | No | `5s` | Timeout for the connection check |
| `-n`, `--namespace` | No | from kubeconfig | Pod namespace |

Standard kubectl flags (`--kubeconfig`, `--context`, etc.) are supported via `genericclioptions`.

## Project structure

- **cmd/plugin** — entrypoint (`main.go`); builds the binary `kubectl-reach`.
- **pkg/reach** — CLI and reach logic (Cobra command, ephemeral container, logs).
- **pkg/version** — version string (set at build time).
- **.github/workflows** — CI workflow (test + lint on push/PR).

## Build & Install

```bash
go mod tidy
make build
# Binary: bin/kubectl-reach — copy to PATH, e.g.:
cp bin/kubectl-reach $(go env GOPATH)/bin/
```

### Makefile

- `make build` / `make bin` — build for current platform into `bin/`.
- `make fmt` / `make vet` / `make test` — format, vet, unit tests.
- `make lint` — run [golangci-lint](https://golangci-lint.run/) (install separately).
- `make ci` — fmt + vet + test + lint (run before push).
- `make verify` — fmt + tidy + vet + test.
- `make build-all` — cross-build for Linux/Darwin/Windows (amd64, arm64).

## Tests and Linting

- Unit tests: `go test ./pkg/... ./cmd/...` or `make test`.
- Linter: `golangci-lint run ./...` or `make lint` (requires [golangci-lint](https://golangci-lint.run/usage/install/)).
- CI (`.github/workflows/ci.yml`) runs test and lint on push/PR to `main` or `master`.

## License

Apache-2.0
