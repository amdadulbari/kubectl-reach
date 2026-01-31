# kubectl-reach

A `kubectl` plugin that tests network connectivity **from** an existing Pod **to** a target (IP, DNS, or Service) by injecting an ephemeral debug container into the source Pod. Useful for verifying if NetworkPolicies or Service Meshes are blocking traffic.

**Requirements:** Kubernetes 1.25+ (EphemeralContainers). Pure Go; no shell dependencies. Linux, macOS, Windows.

---

## Usage

```bash
kubectl reach <source-pod-name> --to <target:port> [options]
```

- **&lt;source-pod-name&gt;** — Pod to run the check from (same network context as your workload).
- **--to &lt;host:port&gt;** — Target to probe (IP, hostname, or Kubernetes Service DNS name).

Use `-n` or `--namespace` for the pod’s namespace; standard kubectl flags (`--context`, `--kubeconfig`) work as usual.

---

## Examples

### Reach a public host (HTTPS)

```bash
kubectl reach myapp-7d4b9c-xk2lm --to google.com:443
```

**Sample output (success):**

```
Connection to google.com 443 port [tcp/https] succeeded!
```

### Reach an in-cluster Service

```bash
kubectl reach frontend-abc123 --to backend:8080 -n myapp
```

**Sample output (success):**

```
Connection to backend 8080 port [tcp/http-alt] succeeded!
```

### Reach an internal IP

```bash
kubectl reach myapp-7d4b9c-xk2lm --to 10.96.0.1:443 -n default
```

**Sample output (success):**

```
Connection to 10.96.0.1 443 port [tcp/https] succeeded!
```

### Connection refused (port closed or blocked)

When the target host is reachable but nothing is listening on the port (or a NetworkPolicy blocks it):

```bash
kubectl reach myapp-7d4b9c-xk2lm --to 10.0.0.5:9999
```

**Sample output:**

```
nc: can't connect to remote host (10.0.0.5): Connection refused
```

### Connection timed out

When the target is unreachable (e.g. wrong network, firewall, or bad host):

```bash
kubectl reach myapp-7d4b9c-xk2lm --to 10.255.255.254:9999 --timeout 3
```

**Sample output:**

```
nc: timeout
```

### Custom debug image and timeout

```bash
kubectl reach myapp-7d4b9c-xk2lm --to myservice:80 --image busybox --timeout 10 -n myns
```

---

## Flags

**Plugin flags**

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--to` | Yes | - | Target as `host:port` (e.g. `google.com:443`, `mysvc:80`) |
| `--image` | No | `busybox` | Debug container image (must provide `nc`) |
| `--timeout` | No | `5s` | Timeout for the connection check |

**Standard kubectl flags** (from `genericclioptions`; same as other kubectl commands)

| Flag | Description |
|------|--------------|
| `-n`, `--namespace` | Pod namespace (default: from kubeconfig) |
| `--context` | Kubernetes context |
| `--kubeconfig` | Path to kubeconfig |
| `--cluster` | Cluster name |
| `--request-timeout` | Request timeout for API server calls |
| `--server`, `-s` | API server address |
| `--user` | User (in kubeconfig) |
| `--token` | Bearer token |
| `--insecure-skip-tls-verify` | Skip TLS verification |

---

## Build & Install

```bash
go mod tidy
make build
cp bin/kubectl-reach $(go env GOPATH)/bin/
```

**Makefile:** `make build` / `make test` / `make lint` / `make ci` / `make build-all` (cross-build). See `make help` or the Makefile for details.

---

## Project structure

- **cmd/plugin** — entrypoint; builds `kubectl-reach`.
- **pkg/reach** — CLI and reach logic (ephemeral container, logs).
- **pkg/version** — version (set at build time).
- **.github/workflows** — CI (test + lint) and release (GoReleaser on tag).

## License

Apache-2.0
