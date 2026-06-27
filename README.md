# Relic Ring Protocol

A distributed Go implementation of the LAUNCH26 Phase 01 challenge. Each planet runs as an independent HTTP service, a Go orchestrator calculates the lowest-latency valid path and manages failures, and a Wails + React desktop control centre visualises the network.

## Features

- Dynamic `universe-config.json` loading and validation
- Official void distance, atmosphere, fibre and tower-delay formulas
- Clockwise tower placement beginning at Tower 0 on the top
- Reversible codex translation and binary framing
- Expanded-state Dijkstra routing with relay tower costs
- Mandatory packet fields and ordered mathematical hop logs
- Independent planet services and Docker Compose deployment
- Node/link chaos controls and rerouting of the next packet
- Wails desktop map, latency panel, route view and live SSE telemetry

## Prerequisites

- Go 1.26 recommended
- Node.js 22+
- Wails CLI 2.12+
- Docker Desktop with Compose
- Git and VS Code

Install Wails:

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
```

## Fastest way to run

```powershell
git clone <your-repository-url>
cd relic-ring-protocol
go mod tidy
cd frontend
npm install
cd ..
docker compose -f deployments/docker/compose.yaml up --build
```

Keep Docker running. Open a second VS Code terminal:

```powershell
wails dev
```

The desktop app connects to `http://localhost:8080`.

## Local backend without Docker

Open seven terminals and start the six planets:

```powershell
go run ./cmd/planet-node --planet Aegis --port 8101
go run ./cmd/planet-node --planet Boreas --port 8102
go run ./cmd/planet-node --planet Dawn --port 8103
go run ./cmd/planet-node --planet Elysium --port 8104
go run ./cmd/planet-node --planet Fenix --port 8105
go run ./cmd/planet-node --planet Caelum --port 8106
```

Then start the orchestrator:

```powershell
go run ./cmd/orchestrator
```

## API examples

```powershell
Invoke-RestMethod http://localhost:8080/api/universe
```

```powershell
$body = @{ origin_id='Aegis'; destination_id='Caelum'; payload='Hello world' } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/transmissions -ContentType 'application/json' -Body $body
```

## Chaos test

```powershell
docker compose -f deployments/docker/compose.yaml stop dawn
```

Wait a few seconds and send the next packet. The route will exclude Dawn. Restore it with:

```powershell
docker compose -f deployments/docker/compose.yaml start dawn
```

## Testing

```powershell
go test ./...
go vet ./...
cd frontend
npm run build
```

## Important implementation details

- Physical constants are read from `configs/universe-config.json`.
- The official void distance uses scaled planet centres and subtracts both radius-plus-atmosphere values.
- Closest tower pairs are used for internal fibre routing and logs, not to replace the official void-distance formula.
- All latency is stored in seconds internally.
- The codex frame joins tokens with commas, converts the frame to UTF-8 and displays each byte as eight binary bits.
- The baseline reroutes the next packet after a failure. In-flight rerouting can be added as an extension.

See `docs/architecture.md`, `docs/assumptions.md` and `docs/vscode-setup.md`.
