# P9E Space — Deployment

This directory contains deployment artefacts for the 17 ConnectRPC
services that make up the platform.

## Layout

```
deploy/
├── docker/                       # Local docker-compose stack
│   ├── docker-compose.yaml
│   └── initdb/
│       └── 00-create-databases.sql
└── helm/
    └── space/                    # Umbrella Kubernetes chart
        ├── Chart.yaml
        ├── values.yaml
        └── templates/
            ├── _helpers.tpl
            ├── configmap.yaml
            ├── deployment.yaml
            ├── service.yaml
            ├── serviceaccount.yaml
            └── NOTES.txt
```

## Local development with docker-compose

Brings up Postgres (with one DB per service), Kafka + Zookeeper, Redis,
and all 17 services.

```bash
cd deploy/docker
docker compose build      # Build all 17 service images
docker compose up -d      # Start the stack
docker compose ps         # Check status
docker compose logs -f eo-catalog   # Tail one service

# Hit a service's health endpoint
curl http://localhost:18000/health  # eo-catalog
```

Service ports (host → container 8080):

| Port  | Service        |
|-------|----------------|
| 18000 | eo-catalog     |
| 18001 | eo-pipeline    |
| 18002 | eo-analytics   |
| 18003 | sat-mission    |
| 18004 | sat-fsw        |
| 18005 | sat-telemetry  |
| 18006 | sat-command    |
| 18007 | sat-simulation |
| 18008 | gs-mc          |
| 18009 | gs-rf          |
| 18010 | gs-scheduler   |
| 18011 | gs-ingest      |
| 18012 | gi-fusion      |
| 18013 | gi-analytics   |
| 18014 | gi-tiles       |
| 18015 | gi-reports     |
| 18016 | gi-predict     |

## Kubernetes with Helm

The `helm/space` chart renders a Deployment + Service + ConfigMap for
each enabled service in `values.yaml`.

```bash
# Create the Postgres password secret first
kubectl create secret generic postgres-credentials \
  --from-literal=password=<your-password> \
  -n p9e-space

# Render templates locally to inspect
helm template my-release deploy/helm/space \
  --set global.imageRegistry=ghcr.io/yourorg \
  --set global.imageTag=0.1.0

# Install
helm install my-release deploy/helm/space \
  --namespace p9e-space --create-namespace \
  --set global.imageRegistry=ghcr.io/yourorg \
  --set global.imageTag=0.1.0

# Disable specific services in values overrides
#   services.gi-predict.enabled: false

# Upgrade
helm upgrade my-release deploy/helm/space --reuse-values

# Uninstall
helm uninstall my-release -n p9e-space
```

Each pod gets:
- `/etc/<svc>/config/config.yaml` mounted from a ConfigMap with non-secret
  defaults (service identity, HTTP/metrics ports, shutdown timeout, CORS).
- `DATABASE_URL` env var assembled from `global.database` + per-service
  database name + the password from the `postgres-credentials` secret.
- Liveness probe on `GET /health`, readiness probe on `GET /ready`.
- Prometheus scrape annotations on port 9090.

## Customising per-environment

Create `values.dev.yaml`, `values.staging.yaml`, etc. and pass with
`-f`:

```yaml
# values.dev.yaml
global:
  cors:
    allowedOrigins: ["http://localhost:3000"]
services:
  sat-telemetry:
    replicas: 1
  gs-ingest:
    replicas: 1
```

```bash
helm install my-release deploy/helm/space \
  -n p9e-space --create-namespace \
  -f deploy/helm/space/values.yaml \
  -f values.dev.yaml
```
