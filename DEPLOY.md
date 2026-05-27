# Deployment Guide

> Deploy MalteGO on a fresh Linux machine in under 5 minutes.

---

## Prerequisites

| Tool | Min version | Check |
|------|-------------|-------|
| Docker | 24+ | `docker --version` |
| Docker Compose | v2 | `docker compose version` |
| Git | any | `git --version` |

---

## 1 — Install Docker

Skip if Docker is already installed.

```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker
```

---

## 2 — Clone the repository

```bash
git clone https://github.com/GeorgeTolen/MalteGO.git
cd MalteGO
```

---

## 3 — Configure

```bash
cp .env.example .env
nano .env
```

**No API key? Use demo mode** — all 13 transforms return realistic fake data:

```env
GREYNOISE_API_KEY=
MOCK_MODE=true
```

**Have a GreyNoise key?** Drop it in and disable mock mode:

```env
GREYNOISE_API_KEY=your_key_here
MOCK_MODE=false
```

> Get a free Community key at [viz.greynoise.io/signup](https://viz.greynoise.io/signup)  
> Enterprise Trial (14 days, all transforms): contact GreyNoise support

---

## 4 — Start

```bash
docker compose up --build -d
```

Wait ~30 seconds for all services to initialize, then verify:

```bash
docker compose ps
```

Expected output — all four containers `healthy`:

```
NAME                       STATUS
maltego-postgres-1         Up (healthy)
maltego-greynoise-api-1    Up (healthy)
maltego-transforms-1       Up (healthy)
maltego-ui-1               Up (healthy)
```

---

## 5 — Open

```
http://localhost:3000
```

Type any IP address, select a transform, click **Run Transform**.

---

## Ports

| Service | Host port | Internal port |
|---------|-----------|---------------|
| Web UI | **3000** | 3000 |
| Transforms API | **8080** | 8080 |
| GreyNoise API | **8090** | 8090 |
| PostgreSQL | **5434** | 5432 |

**Port conflict?** Change the left number in `docker-compose.yml`:

```yaml
ports:
  - "NEW_PORT:OLD_PORT"
```

---

## Common commands

```bash
# Stop all services
docker compose down

# Stop and wipe the database
docker compose down -v

# View live logs
docker compose logs -f

# View logs for one service
docker compose logs -f transforms

# Rebuild after changing .env or code
docker compose up --build -d

# Restart a single service
docker compose restart transforms
```

---

## Switching from demo to live

1. Edit `.env`:
   ```env
   GREYNOISE_API_KEY=your_real_key
   MOCK_MODE=false
   ```
2. Restart:
   ```bash
   docker compose up -d --build
   ```

No data is lost — the graph database persists in the `maltego-data` Docker volume.

---

## Troubleshooting

**Transforms dropdown shows "Loading…"**  
The transforms service isn't reachable from the UI. Check:
```bash
docker compose logs transforms
```

**Save graph fails**  
PostgreSQL might not be running:
```bash
docker compose ps postgres
docker compose restart postgres
```

**Port already in use**  
Find and stop what's using it:
```bash
sudo lsof -i :8080
# or
sudo ss -tlnp | grep 8080
```

**Full reset**
```bash
docker compose down -v
docker compose up --build -d
```
