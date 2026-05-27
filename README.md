# MalteGO

Go port of [greynoise-maltego](https://github.com/GreyNoise-Intelligence/greynoise-maltego) — a threat intelligence graph tool powered by the GreyNoise API.

Take an IP address, run a transform, and the graph expands into actors, CVEs, similar hosts, organizations, and everything connected to that IP. Originally a Python/Flask service, rewritten in Go with a microservices architecture and a built-in web UI — no Maltego Desktop required.

---

## Architecture

```
Browser → UI Service :3000 → Transform Service :8080 → GreyNoise API Service :8090 → api.greynoise.io
```

Three independent services, each in its own Docker container:

| Service | Port | Role |
|---------|------|------|
| `greynoise-api` | 8090 | Wraps GreyNoise API, holds the API key |
| `transforms` | 8080 | 13 transform implementations + Maltego TRX protocol |
| `ui` | 3000 | Web graph interface (Cytoscape.js) |

---

## Transforms

**Community** (free key):
- `GreyNoiseCommunityIPLookup` — noise/riot status, classification (malicious/benign/unknown), actor name

**IP Context** (Enterprise):
- `GreyNoiseNoiseIPLookupAllDetails` — full portrait: geo, ASN, tags, CVEs, ports, actor
- `GreyNoiseNoiseIPLookupGetActor` — threat actor behind the IP
- `GreyNoiseNoiseIPLookupGetCVEs` — vulnerabilities being exploited
- `GreyNoiseNoiseIPLookupGetOrg` — organization and provider
- `GreyNoiseNoiseIPLookupGetPorts` — open/scanning ports
- `GreyNoiseNoiseIPLookupGetTags` — activity tags
- `GreyNoiseRIOTIPLookup` — is this a known legitimate service? (reduces false positives)

**Pivoting** (Enterprise):
- `GreyNoiseNoiseIPSims` — similar IPs by behavior, maps attacker infrastructure

**GNQL Queries** (Enterprise):
- `GreyNoiseQueryByASN` — all malicious IPs in a network
- `GreyNoiseQueryByActor` — all IPs attributed to a threat actor
- `GreyNoiseQueryByCVE` — IPs actively exploiting a vulnerability right now
- `GreyNoiseQueryByTag` — IPs by activity tag

---

## Stack

- **Go 1.26** + **Gin**
- **Cytoscape.js** graph visualization
- **Maltego TRX** XML protocol (backward compatible with Maltego Desktop)
- **GreyNoise v3 API**
- **Docker** + docker-compose

---

## Quick Start

```bash
git clone https://github.com/GeorgeTolen/MalteGO.git
cd MalteGO

cp .env.example .env
# Add your GREYNOISE_API_KEY to .env
# No key yet? Set MOCK_MODE=true for a full demo with realistic fake data
# Free key: https://viz.greynoise.io/signup

# Windows
scripts\run-all.bat

# Docker (Linux / Windows)
docker-compose up --build -d
```

Open **http://localhost:3000** — type an IP, pick a transform, see the graph.

> **Linux deployment guide:** [DEPLOY.md](DEPLOY.md)

---

## API Key

| Tier | Price | Access |
|------|-------|--------|
| Community | Free | 1 transform, 50 requests/week |
| Enterprise Trial | Free (14 days) | All 13 transforms |
| Enterprise | Paid | All 13 transforms, unlimited |

Get a free key: [viz.greynoise.io/signup](https://viz.greynoise.io/signup)

---

## Windows Scripts

```
scripts\run-all.bat      — start all 3 services locally
scripts\run.bat          — start transform service only
scripts\build.bat        — build all 3 binaries into bin\
scripts\test.bat         — run tests
scripts\docker-run.bat   — docker-compose up --build
scripts\docker-stop.bat  — docker-compose down
```

---

## API

Full endpoint reference and Postman examples: [API.md](API.md)

Quick test (Community transform, no Enterprise key needed):

```bash
curl -X POST http://localhost:8080/run/GreyNoiseCommunityIPLookup \
  -H "Content-Type: text/xml" \
  -d '<MaltegoMessage><MaltegoTransformRequestMessage>
    <Entities><Entity Type="maltego.IPv4Address">
      <Value>8.8.8.8</Value><Weight>100</Weight>
    </Entity></Entities>
    <TransformFields>
      <Field Name="greynoise.api.key">YOUR_KEY</Field>
    </TransformFields>
    <Limits SoftLimit="12" HardLimit="12"/>
  </MaltegoTransformRequestMessage></MaltegoMessage>'
```

Web UI JSON API (used by the graph interface):

```bash
curl -X POST http://localhost:8080/api/run/GreyNoiseCommunityIPLookup \
  -H "Content-Type: application/json" \
  -d '{"value": "8.8.8.8"}'
```
