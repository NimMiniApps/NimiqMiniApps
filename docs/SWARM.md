# Docker Swarm Deploy

Postgres uses a local named volume, so it must stay on one node unless you add
shared storage. Pin it by applying the `nimiqminiapps.pgdata=true` label to one
Swarm node.

## Deploy

```bash
docker node update --label-add nimiqminiapps.pgdata=true vm-swarm-worker-01

export POSTGRES_PASSWORD='change-me-url-safe-password'
export ADMIN_TOKEN='change-me'

docker stack deploy -c docker-stack.yml nimiqminiapps
```

The public services are routed by Traefik on the external `web` overlay. The
frontend is static nginx and calls the backend API domain directly:

- `https://nimiqminiapps.com` -> frontend port 80
- `https://api.nimiqminiapps.com` -> backend port 8080

The stack defaults to the GHCR images published by CI:

- `ghcr.io/nimminiapps/nimiq-mini-apps-backend:latest`
- `ghcr.io/nimminiapps/nimiq-mini-apps-frontend:latest`

For a local single-node swarm, build and deploy local images instead:

```bash
docker build -t nimiqminiapps-backend:latest backend
docker build \
  --build-arg VITE_API_BASE_URL=https://api.nimiqminiapps.com \
  -t nimiqminiapps-frontend:latest frontend

export BACKEND_IMAGE=nimiqminiapps-backend:latest
export FRONTEND_IMAGE=nimiqminiapps-frontend:latest
docker stack deploy -c docker-stack.yml nimiqminiapps
```

## Verify

```bash
docker stack services nimiqminiapps
docker service ps nimiqminiapps_postgres
curl -fsS https://nimiqminiapps.com/health
curl -fsS https://api.nimiqminiapps.com/health
```
