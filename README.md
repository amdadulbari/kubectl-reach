# kubectl-reach

A `kubectl` plugin that tests network connectivity **from** an existing Pod **to** a target (IP, DNS, or Service) by injecting an ephemeral debug container into the source Pod. Useful for verifying if NetworkPolicies or Service Meshes are blocking traffic.

- [Krew](https://krew.sigs.k8s.io/) plugin index ready (see [deploy/krew/reach.yaml](deploy/krew/reach.yaml)).
- Project structure follows [krew-plugin-template](https://github.com/replicatedhq/krew-plugin-template) and [Krew best practices](https://krew.sigs.k8s.io/docs/developer-guide/develop/best-practices/).

## Requirements

- **Kubernetes 1.25+** (EphemeralContainers is stable).
- No shell dependencies; pure Go. Works on Linux, macOS, and Windows.

## Usage

```bash
kubectl reach <source-pod-name> --to <target:port> [--namespace <ns>] [options]
```

### Examples

```bash
kubectl reach myapp-abc123 --to google.com:443
kubectl reach myapp-abc123 --to 10.0.0.5:8080 -n myns
kubectl reach myapp-abc123 --to myservice:80 --image busybox --timeout 10
```

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
- **deploy/krew** — Krew manifest ([reach.yaml](deploy/krew/reach.yaml)).
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
- `make dist` — build and create Krew-ready tarballs in `dist/`.

## Krew

1. Fill in URIs and SHA256 in [deploy/krew/reach.yaml](deploy/krew/reach.yaml) (or use GoReleaser release to get URLs).
2. Test locally: `kubectl krew install --manifest=deploy/krew/reach.yaml`
3. Validate: `kubectl krew validate --manifest=deploy/krew/reach.yaml`

## E2E scenarios (k3s / kubectl)

Two live scenarios verify that the tool **detects and reports** reachability vs unreachability:

1. **Reachable:** From a pod, reach `kubernetes.default.svc.cluster.local:443` → tool reports **open** (success).
2. **Unreachable:** From another pod, reach `10.255.255.254:9999` with short timeout → tool reports **Connection timed out** (failure).

See [doc/E2E-SCENARIOS.md](doc/E2E-SCENARIOS.md). Quick run:

```bash
make build
./scripts/e2e-setup.sh      # create namespace + 2 source pods
./scripts/e2e-scenarios.sh  # run both scenarios (PASS/FAIL)
```

## Tests and Linting

- Unit tests: `go test ./pkg/... ./cmd/...` or `make test`.
- Linter: `golangci-lint run ./...` or `make lint` (requires [golangci-lint](https://golangci-lint.run/usage/install/)).
- CI (`.github/workflows/ci.yml`) runs test and lint on push/PR to `main` or `master`.

## License

Apache-2.0
