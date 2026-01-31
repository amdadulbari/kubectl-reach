# kubectl reach — Usage

Test network connectivity **from** a Pod **to** a target (IP, DNS, or Service) using an ephemeral debug container.

## Usage

```bash
kubectl reach <source-pod-name> --to <target:port> [flags]
```

## Examples

```bash
kubectl reach myapp-abc123 --to google.com:443
kubectl reach myapp-abc123 --to 10.0.0.5:8080 -n myns
kubectl reach myapp-abc123 --to myservice:80 --image busybox --timeout 10
```

## Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--to` | Yes | - | Target as `host:port` |
| `--image` | No | `busybox` | Debug container image |
| `--timeout` | No | `5s` | Connection check timeout |
| `-n`, `--namespace` | No | from kubeconfig | Pod namespace |

Standard kubectl flags (`--kubeconfig`, `--context`, etc.) are supported.

## Requirements

- Kubernetes 1.25+ (EphemeralContainers is stable).
