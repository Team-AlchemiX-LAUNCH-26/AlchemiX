# AlchemiX Phase 2 - Relic Ring Protocol

AlchemiX Phase 2 is a Go, Wails, React, and Python ML system for resilient interplanetary packet routing. The application loads a universe topology, computes valid low-latency routes, simulates independent planet nodes, visualizes the network in a desktop control center, and adds an analytical co-pilot that scores routes using congestion, trust, targeting-risk, and uncertainty signals.

## Quick Walkthrough

The project has three main parts:

1. **Distributed routing backend**: Go services in `cmd/`, `internal/orchestrator/`, `internal/planet/`, `internal/routing/`, `internal/latency/`, and `internal/packet/` load the universe, calculate routes, send packets between planet nodes, handle failures, and publish telemetry.
2. **Desktop UI**: A Wails desktop app binds Go methods from `app.go` and `agent_app.go` into the React frontend under `frontend/src/`. The UI shows the universe map, latency panels, decision audit logs, packet timeline, and agent dashboard.
3. **Machine-learning layer**: Python scripts in `training/` train congestion, trust, and targeting models from CSV datasets in `datasets/`. Runtime predictions are served through `training/service/app.py`, and Go calls that service through `internal/intelligence/python_service.go`.

At runtime, a user submits a transmission request. The orchestrator checks node health, builds a graph that excludes failed nodes, links, or towers, calculates a route, creates a packet, sends it to the source planet service, and streams network events back to the desktop UI. The Phase 2 agent separately evaluates candidate next hops using live observations and ML predictions, then records why it continued, rerouted, or queued.

## Repository Structure

| Path | Purpose |
| --- | --- |
| `main.go`, `app.go`, `agent_app.go`, `agent_wiring.go` | Wails desktop entrypoint and Go-to-frontend bindings. |
| `cmd/orchestrator/` | Starts the HTTP orchestrator on port `8080` by default. |
| `cmd/planet-node/` | Starts one independent planet node service. |
| `internal/orchestrator/` | Universe API, health monitoring, route execution, failure controls, SSE events. |
| `internal/routing/` | Graph builder, Dijkstra route search, route details, tower/link validation. |
| `internal/latency/` | Physical latency calculations for void, atmosphere, fiber, tower delay, and route totals. |
| `internal/packet/`, `internal/encoding/` | Packet creation, payload encoding/decoding, binary framing, integrity checks. |
| `internal/agent/` | Phase 2 decision agent, parser, candidate evaluation, policy engine, true-cost calculation. |
| `internal/intelligence/` | Go adapters for Python model predictions. |
| `internal/models/` | Exported model artifacts and manifest used by the runtime and reports. |
| `training/` | Data validation, feature engineering, training, model evaluation, FastAPI prediction service, tests. |
| `datasets/` | Source CSV files and universe configuration for ML and topology. |
| `frontend/src/` | React control center components, services, styles, and TypeScript types. |
| `deployments/docker/` | Docker Compose setup for orchestrator and planet nodes. |
| `tests/`, `internal/*/*_test.go`, `training/tests/` | Go, integration, agent, and ML leakage/preprocessing tests. |
| `docs/` | Architecture notes, assumptions, verification, and sample transmission outputs. |

## Prerequisites

- Go 1.26 recommended
- Node.js 22+
- Wails CLI 2.12+
- Python 3.11+ recommended
- Docker Desktop with Compose

Install Wails:

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
```

Install Python dependencies:

```powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
python -m pip install -r training\requirements.txt
```

Install frontend dependencies:

```powershell
cd frontend
npm install
cd ..
```

## Running the Project

Start the Python ML service first. The Wails app requires it so the runtime uses the exported best-validation `.joblib` models rather than silently falling back to another path.

```powershell
.\.venv\Scripts\Activate.ps1
python -B -m uvicorn training.service.app:app --host 127.0.0.1 --port 8100
```

Start the distributed backend with Docker:

```powershell
docker compose -f deployments/docker/compose.yaml up --build
```

Start the Wails desktop app:

```powershell
wails dev -skipbindings -tags native_webview2loader
```

Useful environment variables:

| Variable | Default | Purpose |
| --- | --- | --- |
| `ML_SERVICE_URL` | `http://127.0.0.1:8100` | Python FastAPI model service used by Go. |
| `ORCHESTRATOR_URL` | `http://localhost:8080` | Desktop app's backend API target. |
| `CHIMERA_API_KEY` | empty | Enables live Chimera API integration in `agent_wiring.go`. |
| `CHIMERA_BASE_URL` | `https://chimera.launch26.space` | Live API base URL. |
| `CONFIG_PATH` | `configs/universe-config.json` | Orchestrator universe config path. |

## Running Without Docker

Start each planet node in separate terminals:

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

## API Examples

Get the current universe snapshot:

```powershell
Invoke-RestMethod http://localhost:8080/api/universe
```

Send a packet:

```powershell
$body = @{ origin_id='Aegis'; destination_id='Caelum'; payload='Hello world' } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/transmissions -ContentType 'application/json' -Body $body
```

Simulate a node failure:

```powershell
docker compose -f deployments/docker/compose.yaml stop dawn
```

Restore it:

```powershell
docker compose -f deployments/docker/compose.yaml start dawn
```

## Machine Learning Workflow

The authoritative ML scripts are in `training/`. The notebooks are useful for inspection and presentation, but reproducible artifacts should come from the scripts.

```powershell
cd training
python validate_data.py
python train_congestion.py
python train_trust.py
python train_targeting.py
python evaluate_all.py
python -m pytest tests
```

Main outputs:

- `internal/models/congestion_model.json`
- `internal/models/trust_model.json`
- `internal/models/targeting_model.json`
- `internal/models/congestion_model.joblib`
- `internal/models/trust_model.joblib`
- `internal/models/targeting_model.joblib`
- `internal/models/model_manifest.json`
- `training/reports/model_selection_summary.csv`
- `training/reports/final_test_metrics.json`

The FastAPI service exposes:

- `GET /health`
- `GET /model-info`
- `POST /predict/congestion`
- `POST /predict/trust`
- `POST /predict/targeting`
- `POST /predict/all`

## Verification

Core verification commands:

```powershell
go test ./...
go vet ./...
cd frontend
npm run build
cd ..
cd training
python -m pytest tests
```

The full Windows verification script is:

```powershell
.\scripts\verify_all.ps1
```

## Developer Report for Judges

Use this section to explain selected procedures with concrete code references.

### How did you manage model training and model selection?

Training is handled by separate scripts for each signal: `training/train_congestion.py`, `training/train_trust.py`, and `training/train_targeting.py`. Each trainer validates schema, splits by time, fits preprocessing only on the training partition, compares multiple model families, records metrics, and exports artifacts.

The key leakage-control decision is in `training/split.py`: `tick_split()` keeps all rows from the same tick together and creates a 70/15/15 train/validation/test split by tick range. `assert_no_overlap()` verifies that train, validation, and test ticks never overlap.

Feature leakage is controlled in `training/features.py`. Rolling and historical features call `shift(1)` before rolling or expanding calculations, so features at tick `t` use only data from earlier ticks. For example, congestion uses previous load ratio, rolling mean load ratio, rolling standard deviation, load change, and distance to saturation without reading the current target.

Model selection is explicit in `training/train_congestion.py`: the congestion trainer compares a median baseline, polynomial ridge, random forest, extra trees, gradient boosting, and HistGradientBoosting. It tracks both `best_validation_model` and `selected_model`. The `.joblib` runtime artifact stores the best validation estimator, while the JSON export keeps compatibility for Go-readable model inspection/export. `training/evaluate_all.py` consolidates metrics into `internal/models/model_manifest.json` and `training/reports/model_selection_summary.csv`.

At runtime, Go does not reimplement sklearn. `internal/intelligence/python_service.go` calls the Python FastAPI service, and `training/service/app.py` loads the `.joblib` models through `PythonModelStore`. This allows the runtime to use the best validation model even when that model is not the legacy Go-compatible HistGradientBoosting JSON export.

### How is the final route decision made?

The agent pipeline begins in `internal/agent/agent.go`. `Execute()` parses the user request, finds a baseline route, generates top-K candidate routes, and then calls the sequential decision loop.

`internal/agent/decision_loop.go` evaluates the route hop by hop. For each possible next hop, it:

- reads the latest network state from the state provider,
- skips invalid links,
- requests congestion, trust, and targeting predictions,
- reads uncertainty for the link,
- calculates a combined score through `TrueCostCalculator`,
- passes viable candidates to the policy engine.

The combined score is implemented in `internal/agent/true_cost.go`:

```text
true_cost =
physical_latency_ms
+ predicted_congestion_penalty_ms
+ trust_weight_ms * (1 - trust_score)
+ targeting_weight_ms * targeting_risk_score
+ uncertainty_weight_ms * uncertainty_score
+ optional_switching_penalty
```

The weights and hard thresholds come from `configs/agent_config.json`. `internal/agent/policy.go` filters out links below the trust threshold or above the targeting-risk threshold, sorts viable hops by combined cost, and returns `CONTINUE`, `REROUTE`, or `QUEUE`.

### How is physical routing implemented?

The baseline route engine is in `internal/routing/`. `internal/routing/dijkstra.go` implements Dijkstra over route states, with deterministic tie-breaking by cost, path length, and lexical path. It skips transitions that become impossible because failed towers block internal ring traversal.

`internal/routing/route_finder.go` turns a path into detailed latency information. It selects entry and exit towers for each planet, computes planet transit, computes void transit between planets, and aggregates the total latency. Physical formulas live under `internal/latency/`.

The orchestrator calls this route engine in `internal/orchestrator/service.go` inside `StartTransmission()`. Before routing, it performs a synchronous health refresh. It then builds a graph that excludes failed nodes, disabled links, and disabled towers, creates the packet, sends it to the source planet node, and verifies the node-reported latency total against the route engine's authoritative total.

### How are failures and rerouting handled?

Network state is maintained by `internal/resilience/`. The orchestrator's `RefreshHealth()` and `StartHealthMonitor()` methods in `internal/orchestrator/service.go` poll planet health endpoints and publish `node.failed` or `node.restored` events. Manual controls for nodes, links, and towers are exposed through orchestrator methods such as `DisableNode()`, `DisableLink()`, and `DisableTower()`.

The current route is calculated at transmission time using the latest graph snapshot. Failed nodes and links are excluded before route calculation, so the next packet avoids unavailable infrastructure. Tower failures are also included in graph construction, and candidate transitions are skipped if tower-level traversal is impossible.

### How does the desktop UI connect to Go?

`main.go` starts the Wails app and binds two Go objects: `App` for the distributed network backend and `AgentApp` for Phase 2 agent features.

`app.go` wraps orchestrator HTTP calls for the React UI. Methods like `GetUniverse()`, `StartTransmission()`, `DisableNode()`, `DisableLink()`, and `ResetNetwork()` are callable from the frontend through generated Wails bindings.

`agent_app.go` defines DTOs and frontend-safe methods for the agent: `ParseTransmissionRequest()`, `EvaluateTransmission()`, `GetAgentState()`, `GetDecisionAudit()`, `GetPacketTimeline()`, and `ResetAgentHistory()`. `agent_wiring.go` constructs the real internal agent, attaches Python model clients, wires live state, streams hop events with `runtime.EventsEmit()`, and records audit/timeline entries.

### How is packet integrity handled?

Payload conversion is implemented under `internal/encoding/`. `EncodePayload()` converts text bytes into codex-specific tokens and serializes them. `DecodePayload()` reverses the process, and `DecodeBinary()` handles deserialization plus decoding.

Packet integrity is checked in `internal/packet/processor.go` through `VerifyPayload()`, which fails if the decoded payload differs from the original payload.

## Notes and Assumptions

- Physical latency is stored internally in seconds in the route engine and displayed as needed in the UI.
- The agent's ML-facing cost calculation uses milliseconds because the model features and penalties are expressed in milliseconds.
- Saturated links are handled by hard rules in the ML service and agent configuration rather than relying only on a learned classifier.
- The React frontend is visualization/control logic only; official routing and latency calculations remain in Go.
- Live Chimera API integration is optional. Without `CHIMERA_API_KEY`, the system can still operate with local/mock state.
- See `docs/architecture.md`, `docs/assumptions.md`, `docs/verification.md`, `ML_PIPELINE_README.md`, and `training/README_ML_IMPLEMENTATION.md` for supporting details.
