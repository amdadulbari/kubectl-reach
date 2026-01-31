# How kubectl reach works

## What happens when you run `kubectl reach`?

When you run:

```bash
kubectl reach <pod-name> --to <host:port> -n <namespace>
```

the plugin does **not** run a command inside an existing container. It **adds a new container** to the same pod and runs the connectivity check there.

1. **Get the pod** — The plugin reads the pod spec from the API.
2. **Add an ephemeral container** — It calls the Kubernetes *ephemeralcontainers* subresource to add a temporary container (by default `busybox`) that runs `nc -zv <host> <port>`.
3. **Wait for it to start** — It waits until the ephemeral container appears in the pod status (Running or Terminated).
4. **Stream logs** — It streams that container’s stdout/stderr to your terminal (e.g. “open” or “Connection timed out”).

So **yes**: another container is created **inside the same pod**. That container is an **ephemeral container**: it shares the pod’s network (and other namespaces) but is not a “main” container.

---

## Why does `kubectl get po` still show 1/1 ready?

Pod readiness in Kubernetes is based only on **main containers** (`spec.containers`), not on ephemeral containers.

- **Ready 1/1** means: “1 main container, and that 1 main container is ready.”
- Ephemeral containers:
  - Are **not** counted in the “X/Y” container readiness.
  - Do **not** affect the pod’s `Ready` condition.
  - Are listed separately in the pod status (e.g. `status.ephemeralContainerStatuses`).

So after running `kubectl reach`, you still see **1/1** because the **main** container is unchanged and still the only one that counts for readiness. The ephemeral container exists but is ignored for the “1/1” and for `Ready`.

To see ephemeral containers:

```bash
kubectl get pod <pod-name> -n <namespace> -o jsonpath='{.status.ephemeralContainerStatuses}' | jq .
```

or:

```bash
kubectl get pod <pod-name> -n <namespace> -o yaml
# look at status.ephemeralContainerStatuses
```

---

## After testing once, how can I test again with the same pod?

The plugin uses a **unique ephemeral container name per run** (e.g. `reach-debug-1738234567890123456`). So each run adds a **new** ephemeral container, and you can run `kubectl reach` multiple times on the same pod:

```bash
kubectl reach mypod --to host1:443 -n default
kubectl reach mypod --to host2:80 -n default   # same pod, works
```

- **Downside:** The pod accumulates one ephemeral container per run (they cannot be removed until the pod is deleted).
- **Upside:** You can re-use the same pod for many reach tests without recreating it.

---

## Summary

| Question | Answer |
|----------|--------|
| Does it create another container in the pod? | Yes. An **ephemeral** container is added to the same pod. |
| Why is the pod still 1/1 ready? | Readiness counts only **main** containers; ephemeral containers are not included. |
| How to test again on the same pod? | Just run `kubectl reach` again; each run uses a **unique** ephemeral container name (`reach-debug-<timestamp>`). |
