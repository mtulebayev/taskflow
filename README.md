# taskflow

[![pipeline status](https://gitlab.com/mtulebayev/taskflow/badges/main/pipeline.svg)](https://gitlab.com/mtulebayev/taskflow/-/pipelines)

Task management REST API in Go with a full delivery pipeline:
**GitLab CI → Container Registry → Helm → Kubernetes**.
Every push to `main` is linted, tested, built into an image tagged with the commit SHA and deployed to the cluster automatically.

**Stack:** Go · PostgreSQL · Docker · GitLab CI (self-hosted runner) · Kubernetes (kind) · Helm · Traefik · goose migrations · Prometheus metrics · VictoriaMetrics · Grafana · vmalert (victoria-metrics-k8s-stack)

## Architecture

```mermaid
flowchart LR
  push[git push to main] --> ci
  subgraph ci[GitLab CI on self-hosted runner]
    lint[lint: gofmt, go vet] --> test[test: go test -race] --> build[build: docker image :sha]
  end
  build --> reg[(GitLab Container Registry)]
  build --> deploy[deploy: helm upgrade --install]
  deploy --> cluster
  subgraph cluster[Kubernetes cluster]
    traefik[Traefik ingress] --> app[taskflow x2]
    hook[migrate Job<br/>pre-upgrade hook] --> pg
    app --> pg[(PostgreSQL StatefulSet)]
  end
  client[client] --> traefik
```

## API

| Method | Path | Description |
|---|---|---|
| `GET` | `/tasks` | list tasks |
| `POST` | `/tasks` | create a task: `{"title": "..."}` |
| `GET` | `/tasks/{id}` | get a task |
| `GET` | `/healthz` | liveness: the process is alive |
| `GET` | `/readyz` | readiness: the database is reachable |
| `GET` | `/metrics` | Prometheus metrics: request count and latency by route |

## Run locally

**Docker Compose** (Postgres → migrations → app):

```bash
docker compose up --build
curl http://localhost:8080/tasks
```

**Kubernetes (kind):**

```bash
kind create cluster --config deploy/kind/cluster.yaml
helm repo add traefik https://traefik.github.io/charts
helm install traefik traefik/traefik -n traefik --create-namespace \
  --set service.spec.type=NodePort --set ports.web.nodePort=30080

kubectl create namespace taskflow
kubectl -n taskflow create secret generic postgres-credentials \
  --from-literal=POSTGRES_USER=postgres --from-literal=POSTGRES_PASSWORD=postgres
kubectl -n taskflow create secret generic taskflow-db \
  --from-literal=DATABASE_URL='postgres://postgres:postgres@postgres:5432/taskflow?sslmode=disable'
kubectl apply -f deploy/k8s/postgres.yaml

helm install taskflow deploy/helm/taskflow -n taskflow --set image.tag=<commit-sha> --wait
curl http://taskflow.localhost/tasks
```

Demo data: `scripts/seed.sql` (60 sample tasks).

## CI/CD

| Stage | What it does | Runs on |
|---|---|---|
| lint | `gofmt` check, `go vet` | every branch |
| test | `go test -race` | every branch |
| build | multi-stage image pushed as `:<short-sha>` and `:latest` | `main` |
| deploy | `helm upgrade --install` with the commit SHA | `main` |

The runner is self-hosted (WSL2, docker executor); shared runners are disabled.

## Monitoring

The app exposes RED metrics (rate, errors, duration) on `/metrics`.
A `VMServiceScrape` collects them from every pod; Grafana and alert rules ship
inside the Helm chart, so they are versioned and deployed together with the app.

![Taskflow RED dashboard](docs/dashboard.png)

| Alert | Condition | Severity |
|---|---|---|
| `TaskflowHigh5xxRate` | 5xx > 5% of requests for 5m | warning |
| `TaskflowNoReadyPods` | no available pods for 1m | critical |

Install the stack once per cluster:

```bash
helm repo add vm https://victoriametrics.github.io/helm-charts/
helm install vmks vm/victoria-metrics-k8s-stack -n monitoring --create-namespace \
  -f deploy/monitoring/vm-stack-values.yaml --wait --timeout 10m
```

Grafana: http://grafana.localhost (dashboard **Taskflow / RED**).

## Design decisions

- **Image tag = commit SHA, required by the chart.** No implicit `latest`, every running pod is traceable to a commit via the `app.kubernetes.io/version` label.
- **Migrations are a separate binary** in the same image, run as a Helm `pre-upgrade` hook. A failed migration stops the release before the Deployment changes. Tested: a broken DB password failed the hook while the old pods kept serving traffic, then `helm rollback`.
- **Liveness vs readiness.** `/healthz` only checks the process, so a DB outage does not restart pods; `/readyz` pings the DB and takes the pod out of the Service instead.
- **Graceful shutdown** on SIGTERM with a 10s timeout, within `terminationGracePeriodSeconds`.
- **Minimal, non-root image.** Distroless `static:nonroot`, numeric UID 65532 so `runAsNonRoot` can be verified, read-only root filesystem, no privilege escalation. Image size: ~1.7 GB build stage → ~40 MB final (two static binaries on distroless).
- **Least privilege for CI.** Deploys use a dedicated ServiceAccount bound to the `edit` role only in the `taskflow` namespace; the kubeconfig is a protected file variable scoped to the deploy environment.
- **Traefik instead of ingress-nginx**, which was retired in March 2026.
- **Postgres is outside the app chart.** The app release can be uninstalled without touching the database.
- **Trade-off:** the runner mounts the host Docker socket to build images. Acceptable for a personal runner; on shared runners I would use Kaniko or Buildah.
- **Dashboard as code.** The Grafana dashboard JSON lives in the chart and is loaded by the
  Grafana sidecar from a labelled ConfigMap. Deleting it in the UI does nothing: it comes back from git.
- **Latency buckets tuned for sub-5ms responses.** Default Prometheus buckets start at 5ms, so p95
  was stuck at an interpolation artefact of 4.75ms. Buckets now start at 0.5ms; measured p95 is
  ~1ms for `GET /tasks/{id}` and ~1.2ms for `GET /tasks` (full list, no pagination yet).
- **Probes are not logged.** `/healthz`, `/readyz` and `/metrics` are still counted in metrics
  but excluded from access logs, which would otherwise be mostly kubelet and scraper noise.

## Troubleshooting

Real problems hit while building this and how they were solved: [docs/troubleshooting.md](docs/troubleshooting.md).

## Roadmap

- [ ] Centralized logs (VictoriaLogs / Loki)
- [ ] Ansible playbook provisioning a k3s node
- [ ] Keycloak (OIDC) login for Grafana