# AlchemiX Phase 2 Project Report README

This document is a detailed report README for outside observers, judges, and developers who need to understand the full repository structure and explain selected implementation procedures from code. The root `README.md` is the quick operating guide; this file is the expanded technical map.

## 1. Project Summary

AlchemiX Phase 2, also named Relic Ring Protocol in the codebase, is a resilient interplanetary packet-routing system. It combines:

- A Go distributed backend for universe loading, physical route calculation, packet forwarding, node health, and failure simulation.
- A Wails desktop shell that exposes Go methods directly to a React frontend.
- A React control center for visualizing the universe, routes, latency, hop logs, audit records, and Phase 2 agent decisions.
- A Python machine-learning pipeline that trains congestion, trust, and targeting-risk models from provided datasets.
- A Python FastAPI model service that serves the best validation `.joblib` models to the Go runtime.

The system supports two related demonstrations:

1. **Baseline distributed transmission**: the orchestrator finds the lowest-latency valid route and sends the packet through independent planet node services.
2. **Phase 2 analytical co-pilot**: an agent evaluates possible next hops using physical latency, congestion prediction, trust score, targeting-risk score, uncertainty, and policy thresholds.

## 2. Runtime Flow

### 2.1 Distributed Transmission Flow

1. The Wails app starts in `main.go`.
2. `app.go` creates a remote client for the orchestrator, defaulting to `http://localhost:8080`.
3. The user starts a transmission from the React UI.
4. The frontend calls Wails-generated bindings in `frontend/wailsjs/go/main/App.*`.
5. `app.go` sends the request to the orchestrator API.
6. `internal/orchestrator/service.go` refreshes planet health and builds the current graph.
7. `internal/routing/dijkstra.go` finds a valid lowest-latency path.
8. `internal/routing/route_finder.go` builds physical latency details for the selected path.
9. `internal/packet/factory.go` creates a packet and encoded payload.
10. The orchestrator sends the packet to the source planet node.
11. Planet node handlers forward the packet through the route.
12. The orchestrator streams events through `internal/transport/event_hub.go`.
13. The Wails app relays events into the frontend with `runtime.EventsEmit`.

### 2.2 Phase 2 Agent Flow

1. `agent_wiring.go` loads agent configuration and topology.
2. It verifies that the Python ML service is reachable.
3. `internal/intelligence/python_service.go` creates Go adapters for congestion, trust, and targeting models.
4. `internal/liveapi/client.go` provides live or mock link observations.
5. `internal/candidate/top_k.go` generates candidate routes.
6. `internal/agent/agent.go` parses the user request and starts the decision loop.
7. `internal/agent/decision_loop.go` evaluates each possible next hop.
8. `internal/agent/true_cost.go` combines physical and ML costs.
9. `internal/agent/policy.go` applies hard safety thresholds and selects an action.
10. `agent_app.go` returns a frontend-safe DTO report.
11. React components display the decision report, audit trail, and timeline.

## 3. Procedure Explanations for Judges

### 3.1 Model Training and Selection

The ML pipeline is script-driven and reproducible. The important code is in:

- `training/validate_data.py`
- `training/split.py`
- `training/preprocessing.py`
- `training/features.py`
- `training/train_congestion.py`
- `training/train_trust.py`
- `training/train_targeting.py`
- `training/evaluate_all.py`
- `training/service/app.py`
- `internal/intelligence/python_service.go`

The training process follows this pattern:

1. Validate the raw dataset schema before training.
2. Split rows by tick using `training/split.py`.
3. Fit preprocessing only on the training partition.
4. Build lagged and rolling features inside each partition separately.
5. Compare multiple model families on validation data.
6. Export JSON artifacts for inspection and compatibility.
7. Export `.joblib` artifacts for the active Python runtime service.
8. Write model metadata and summary reports.

The leakage prevention is explicit:

- `tick_split()` in `training/split.py` keeps all rows from the same tick in the same partition.
- `assert_no_overlap()` proves train, validation, and test ticks do not overlap.
- `training/features.py` uses `shift(1)` before rolling or expanding calculations, so a feature at tick `t` only uses earlier observations.

The congestion trainer, `training/train_congestion.py`, compares these model families:

- Median baseline
- Polynomial Ridge
- Random Forest
- Extra Trees
- Gradient Boosting
- HistGradientBoosting

It records both `best_validation_model` and `selected_model`. The runtime `.joblib` model can use the best validation estimator, while JSON exports remain available for Go-compatible inspection. `training/evaluate_all.py` consolidates artifacts into `internal/models/model_manifest.json` and report files in `training/reports/`.

### 3.2 True Cost and Route Decision

The Phase 2 agent decision is based on true cost. The formula is implemented in `internal/agent/true_cost.go`:

```text
true_cost =
physical_latency_ms
+ predicted_congestion_penalty_ms
+ trust_weight_ms * (1 - trust_score)
+ targeting_weight_ms * targeting_risk_score
+ uncertainty_weight_ms * uncertainty_score
+ optional_switching_penalty
```

The weights and thresholds come from `configs/agent_config.json`.

The hop-level process is in `internal/agent/decision_loop.go`:

1. Fetch the latest network state.
2. Find valid next hops from candidate routes.
3. Skip invalid or saturated links.
4. Run congestion, trust, and targeting model adapters.
5. Add uncertainty.
6. Calculate true cost.
7. Pass the scored hops to `internal/agent/policy.go`.

`internal/agent/policy.go` filters candidates using hard thresholds:

- reject if trust score is below `trust_hard_threshold`;
- reject if targeting risk is above `risk_hard_threshold`;
- choose the lowest combined-cost candidate among viable hops;
- queue or reroute when no viable hop exists.

### 3.3 Physical Routing

Physical routing is implemented in the Go backend, not the frontend.

- `internal/routing/graph_builder.go` builds the graph from the universe config and current failures.
- `internal/routing/dijkstra.go` implements lowest-cost route search with deterministic tie-breaking.
- `internal/routing/route_cost.go` computes transition costs between route states.
- `internal/routing/route_finder.go` converts a path into detailed latency breakdowns.
- `internal/latency/*.go` contains the formulas for void, atmosphere, fiber, tower delay, planet transit, and total route latency.
- `internal/geometry/*.go` contains coordinate scaling, tower placement, ring-segment math, closest tower pairs, and void distance helpers.

The orchestrator calls the route engine in `internal/orchestrator/service.go`, especially in `StartTransmission()`.

### 3.4 Failure Handling

Failure state is centralized in `internal/resilience/network_state.go`.

The orchestrator can:

- mark nodes failed or restored;
- disable or enable links;
- disable or enable individual towers;
- rebuild the routing graph with unavailable infrastructure removed;
- publish events such as `node.failed`, `node.restored`, `link.failed`, `tower.failed`, and `network.reset`.

Routes are recalculated for new transmissions against the current graph snapshot. Tower failures can make an internal ring path impossible, and `internal/routing/dijkstra.go` skips those transitions rather than failing the entire search.

## 4. How to Run

### 4.1 Install Dependencies

```powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
python -m pip install -r training\requirements.txt

cd frontend
npm install
cd ..

go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
```

### 4.2 Start the ML Service

```powershell
.\.venv\Scripts\Activate.ps1
python -B -m uvicorn training.service.app:app --host 127.0.0.1 --port 8100
```

### 4.3 Start the Distributed Network

```powershell
docker compose -f deployments/docker/compose.yaml up --build
```

### 4.4 Start the Desktop App

```powershell
wails dev -skipbindings -tags native_webview2loader
```

## 5. Verification Commands

```powershell
go test ./...
go vet ./...

cd frontend
npm run build
cd ..

cd training
python -m pytest tests
cd ..
```

Full Windows verification:

```powershell
.\scripts\verify_all.ps1
```

## 6. Full File Structure and File Responsibilities

The list below follows the live repository structure.

### 6.1 Root Files

| File | Purpose |
| --- | --- |
| `.gitignore` | Git ignore rules for generated files, build output, local environments, and temporary artifacts. |
| `agent_app.go` | Wails-facing adapter for Phase 2 agent functions. Defines DTOs and methods such as `EvaluateTransmission`, `GetDecisionAudit`, and `GetPacketTimeline`. |
| `agent_wiring.go` | Constructs the real agent, loads topology, connects Python ML clients, attaches live state, emits hop events, and records audit/timeline data. |
| `app.go` | Wails-facing adapter for orchestrator/network operations such as universe loading, transmission, failure controls, and reset. |
| `go.mod` | Go module definition and dependency declarations. |
| `go.sum` | Checksums for Go module dependency integrity. |
| `main.go` | Wails desktop entrypoint. Loads `.env`, creates `App` and `AgentApp`, binds them to the frontend, and starts the desktop window. |
| `Makefile` | Convenience commands for formatting, testing, frontend build, Wails dev, and Docker Compose. |
| `README.md` | Main quick-start project README. |
| `PROJECT_REPORT_README.md` | This detailed report README. |
| `PROJECT_COMMANDS.md` | Practical command reference for ML, frontend, Go, Wails, and Docker workflows. |
| `PROJECT_TREE.txt` | Older static tree snapshot kept as a reference. The live tree is more complete than this file. |
| `ML_PIPELINE_README.md` | Detailed ML pipeline guide and operating notes. |
| `GEOSPATIAL_ANALYTICS_UI_TEMPLATE_README.md` | UI template/reference documentation retained with the project. |
| `VS_CODE_SETUP.md` | VS Code setup instructions at repository root. |
| `wails.json` | Wails project configuration, including frontend build settings and app metadata. |

### 6.2 `cmd/`

| File | Purpose |
| --- | --- |
| `cmd/orchestrator/main.go` | Starts the orchestrator HTTP service, loads config, sets server address, starts health monitoring, and handles graceful shutdown. |
| `cmd/planet-node/main.go` | Starts a single planet-node HTTP service for one configured planet and port. |

### 6.3 `configs/`

| File | Purpose |
| --- | --- |
| `configs/agent_config.json` | Phase 2 agent weights, thresholds, retry settings, route settings, and audit buffer size. |
| `configs/example-universe-config.json` | Example universe topology and metadata for reference or template use. |
| `configs/universe-config.json` | Runtime universe configuration used by the orchestrator, route engine, and local services. |

### 6.4 `datasets/`

| File | Purpose |
| --- | --- |
| `datasets/DATASETS.md` | Documentation for the dataset files and expected columns. |
| `datasets/link_incident_history.csv` | Historical incident/jamming labels used mainly by the targeting-risk model. |
| `datasets/link_telemetry.csv` | Historical telemetry and self-reported latency data used mainly by the trust model. |
| `datasets/link_traffic_history.csv` | Historical traffic, load, and observed latency data used mainly by the congestion model. |
| `datasets/universe-config.json` | Dataset-side universe topology used by ML preprocessing and model feature metadata. |

### 6.5 `deployments/`

| File | Purpose |
| --- | --- |
| `deployments/docker/compose.yaml` | Docker Compose network for orchestrator and planet-node services. |
| `deployments/docker/Dockerfile.orchestrator` | Container build recipe for the orchestrator service. |
| `deployments/docker/Dockerfile.planet` | Container build recipe for planet-node services. |
| `deployments/scripts/kill-node.ps1` | Helper script to stop or simulate failure of a node. |
| `deployments/scripts/restore-node.ps1` | Helper script to restore a stopped or failed node. |
| `deployments/scripts/start-network.ps1` | Helper script to start the local distributed network. |
| `deployments/scripts/stop-network.ps1` | Helper script to stop the local distributed network. |

### 6.6 `docs/`

| File | Purpose |
| --- | --- |
| `docs/architecture.md` | Short architecture summary of desktop, orchestrator, route engine, packet flow, and telemetry. |
| `docs/assumptions.md` | Project assumptions and interpretation decisions. |
| `docs/sample-baseline-transmission.json` | Example output for a normal baseline transmission. |
| `docs/sample-hard-failure-transmission.json` | Example output for a transmission under hard failure conditions. |
| `docs/verification.md` | Verification notes and expected validation workflow. |
| `docs/vscode-setup.md` | VS Code environment setup guide. |

### 6.7 `frontend/`

| File | Purpose |
| --- | --- |
| `frontend/index.html` | Vite HTML entrypoint for the React app. |
| `frontend/package.json` | Frontend package metadata, scripts, and dependencies. |
| `frontend/package-lock.json` | Locked npm dependency tree for reproducible frontend installs. |
| `frontend/package.json.md5` | Wails-generated checksum used to detect frontend package changes. |
| `frontend/tsconfig.json` | TypeScript compiler configuration. |
| `frontend/vite.config.ts` | Vite configuration for React development and production builds. |

#### `frontend/src/`

| File | Purpose |
| --- | --- |
| `frontend/src/App.tsx` | Main React application layout and state orchestration for network and agent views. |
| `frontend/src/main.tsx` | React entrypoint that mounts the app into the DOM. |
| `frontend/src/styles.css` | Main stylesheet for the desktop UI. |
| `frontend/src/services/agentApi.ts` | Small wrapper around Wails-generated `AgentApp` methods, isolating generated import paths from the rest of the frontend. |
| `frontend/src/styles/agent.css` | Agent-specific UI styling. |
| `frontend/src/types/agent.ts` | TypeScript types for agent state, parsed requests, decisions, audit records, and packet timeline entries. |
| `frontend/src/types/index.ts` | Shared TypeScript types for universe, route, latency, packets, and network state. |

#### `frontend/src/components/`

| File | Purpose |
| --- | --- |
| `frontend/src/components/AgentDashboard.tsx` | High-level Phase 2 agent dashboard that composes agent status, decisions, audit, and timeline information. |
| `frontend/src/components/AgentPanel.tsx` | Agent interaction panel for parsing/evaluating natural-language transmission requests. |
| `frontend/src/components/ConversionHistory.tsx` | Displays payload conversion or encoding history in the UI. |
| `frontend/src/components/DecisionAuditLog.tsx` | Shows audit records for agent decisions, including cost components and reasons. |
| `frontend/src/components/HopLogTable.tsx` | Displays hop-by-hop packet log details. |
| `frontend/src/components/LatencyPanel.tsx` | Displays route latency totals and breakdowns. |
| `frontend/src/components/LinkEvaluationTable.tsx` | Shows per-link agent evaluation scores such as congestion, trust, risk, uncertainty, and combined cost. |
| `frontend/src/components/PacketTimeline.tsx` | Displays packet lifecycle/timeline entries from the agent adapter. |
| `frontend/src/components/TowerInspector.tsx` | UI component for inspecting and controlling tower-level state. |
| `frontend/src/components/UniverseMap.tsx` | Visual map of planets, links, selected routes, and network status. |

#### `frontend/public/planets/`

| File | Purpose |
| --- | --- |
| `frontend/public/planets/planet-cratered.webp` | Planet texture/image asset for the UI. |
| `frontend/public/planets/planet-desert.webp` | Planet texture/image asset for the UI. |
| `frontend/public/planets/planet-gas-giant.webp` | Planet texture/image asset for the UI. |
| `frontend/public/planets/planet-ice.webp` | Planet texture/image asset for the UI. |
| `frontend/public/planets/planet-ocean.webp` | Planet texture/image asset for the UI. |
| `frontend/public/planets/planet-volcanic.webp` | Planet texture/image asset for the UI. |

#### `frontend/wailsjs/`

These files are generated by Wails and should normally be regenerated, not manually edited.

| File | Purpose |
| --- | --- |
| `frontend/wailsjs/go/main/AgentApp.d.ts` | TypeScript declarations for Wails-exposed `AgentApp` methods. |
| `frontend/wailsjs/go/main/AgentApp.js` | JavaScript bridge for Wails-exposed `AgentApp` methods. |
| `frontend/wailsjs/go/main/App.d.ts` | TypeScript declarations for Wails-exposed `App` methods. |
| `frontend/wailsjs/go/main/App.js` | JavaScript bridge for Wails-exposed `App` methods. |
| `frontend/wailsjs/go/models.ts` | Generated TypeScript models for Go structs exposed to the frontend. |
| `frontend/wailsjs/runtime/package.json` | Runtime package metadata for Wails-generated frontend runtime helpers. |
| `frontend/wailsjs/runtime/runtime.d.ts` | TypeScript declarations for Wails runtime APIs. |
| `frontend/wailsjs/runtime/runtime.js` | JavaScript implementation of Wails runtime APIs used by the frontend. |

### 6.8 `training/`

| File | Purpose |
| --- | --- |
| `training/README_ML_IMPLEMENTATION.md` | ML implementation guide describing delivered signals, workflow, model families, safety rules, and service endpoints. |
| `training/requirements.txt` | Python dependency list for validation, training, FastAPI serving, testing, and notebooks. |
| `training/validate_data.py` | Validates source dataset schemas and expected data constraints before training. |
| `training/validate_models.py` | Validates exported model artifacts and metadata before report/manifest generation. |
| `training/preprocessing.py` | Shared preprocessing helpers for schema validation, universe loading, baseline/capacity calculation, link ID encoding, missing-value handling, and dataset hashing. |
| `training/split.py` | Tick-based train/validation/test splitter with no-overlap assertions. |
| `training/features.py` | Leakage-safe feature engineering for congestion, trust, and targeting models. |
| `training/model_metadata.py` | Shared metadata builder for model artifacts, including versioning, split info, dataset hash, and preprocessing metadata. |
| `training/train_congestion.py` | Trains congestion penalty regressors, compares model families, applies hard saturation rules, and exports JSON/joblib artifacts. |
| `training/train_trust.py` | Trains trust/under-reporting model, evaluates spoof-related metrics, and exports artifacts. |
| `training/train_targeting.py` | Trains targeting-risk classifier from traffic-share and incident history and exports artifacts. |
| `training/evaluate_all.py` | Validates model exports, writes `model_manifest.json`, and creates consolidated report files. |
| `training/export_models.py` | Export utility for model artifacts and compatibility formats. |
| `training/model_validation.json` | Machine-readable validation result summary for trained models. |
| `training/model_validation.md` | Human-readable validation result summary for trained models. |

#### `training/service/`

| File | Purpose |
| --- | --- |
| `training/service/__init__.py` | Marks the service directory as a Python package. |
| `training/service/app.py` | FastAPI prediction service exposing health, model info, and prediction endpoints for congestion, trust, targeting, and combined predictions. |
| `training/service/feature_builder.py` | Builds live prediction features from Go `LinkObservation`-style inputs. |
| `training/service/model_loader.py` | Loads `.joblib` model artifacts and supporting metadata from `internal/models/`. |
| `training/service/schemas.py` | Pydantic request and response schemas for the FastAPI prediction API. |

#### `training/tests/`

| File | Purpose |
| --- | --- |
| `training/tests/test_no_leakage.py` | Tests that feature engineering and split behavior do not leak future information. |
| `training/tests/test_preprocessing.py` | Tests preprocessing logic, schema handling, missing values, and derived fields. |
| `training/tests/test_split.py` | Tests tick-based train/validation/test split behavior and overlap prevention. |

#### `training/notebooks/`

| File | Purpose |
| --- | --- |
| `training/notebooks/01_EDA.ipynb` | Exploratory data analysis notebook. |
| `training/notebooks/02_Data_Preprocessing.ipynb` | Notebook for inspecting preprocessing behavior. |
| `training/notebooks/03_Feature_Engineering.ipynb` | Notebook for inspecting engineered features. |
| `training/notebooks/04_Model_Training.ipynb` | Notebook for model training exploration and presentation. |
| `training/notebooks/05_Model_Evaluation.ipynb` | Notebook for model evaluation analysis. |
| `training/notebooks/06_Model_Export.ipynb` | Notebook for reviewing export artifacts. |

#### `training/reports/`

| File | Purpose |
| --- | --- |
| `training/reports/final_test_metrics.json` | Consolidated model metric report generated by `evaluate_all.py`. |
| `training/reports/model_selection_summary.csv` | CSV summary of selected and best-validation model choices. |

### 6.9 `internal/agent/`

| File | Purpose |
| --- | --- |
| `internal/agent/agent.go` | Top-level Phase 2 agent object and `Execute()` pipeline. |
| `internal/agent/decision_loop.go` | Hop-by-hop decision loop that evaluates candidates using live state and model outputs. |
| `internal/agent/explanation.go` | Builds public explanations for decisions and reports. |
| `internal/agent/parser.go` | Deterministic parser for natural-language transmission requests. |
| `internal/agent/policy.go` | Policy engine that applies trust/risk thresholds and chooses continue, reroute, or queue. |
| `internal/agent/router_adapter.go` | Adapter that wraps route finding for agent candidate/baseline route use. |
| `internal/agent/true_cost.go` | Implements true-cost calculation from latency and ML-derived penalties. |
| `internal/agent/types.go` | Shared agent types, interfaces, observations, predictions, route candidates, actions, and report structs. |

### 6.10 `internal/audit/`

| File | Purpose |
| --- | --- |
| `internal/audit/logger.go` | In-memory audit logger with bounded storage. |
| `internal/audit/record.go` | Audit record struct and constructor for decision evidence. |
| `internal/audit/report.go` | Helpers for audit report formatting or aggregation. |

### 6.11 `internal/candidate/`

| File | Purpose |
| --- | --- |
| `internal/candidate/diversity.go` | Candidate-route diversity helpers for avoiding over-reliance on similar routes. |
| `internal/candidate/recent_usage.go` | Tracks recent route/link usage for selection context. |
| `internal/candidate/top_k.go` | Implements Yen-style K-shortest candidate route generation using physical link weights. |

### 6.12 `internal/config/`

| File | Purpose |
| --- | --- |
| `internal/config/defaults.go` | Default configuration values and fallback behavior. |
| `internal/config/loader.go` | Loads universe configuration JSON into domain structs. |
| `internal/config/loader_test.go` | Tests config loading behavior. |
| `internal/config/planet_lookup.go` | Lookup helper for finding planets by ID. |
| `internal/config/validator.go` | Validates universe config consistency, required fields, links, and values. |
| `internal/config/validator_test.go` | Tests config validation rules. |

### 6.13 `internal/domain/`

| File | Purpose |
| --- | --- |
| `internal/domain/hop_log.go` | Domain structs for hop-level logs. |
| `internal/domain/network_state.go` | Domain structs for network snapshots and planet/link status. |
| `internal/domain/packet.go` | Packet domain model. |
| `internal/domain/planet.go` | Planet domain model. |
| `internal/domain/route.go` | Route and latency breakdown domain models. |
| `internal/domain/tower.go` | Tower-related domain model definitions. |
| `internal/domain/universe.go` | Universe configuration and metadata domain models. |

### 6.14 `internal/encoding/`

| File | Purpose |
| --- | --- |
| `internal/encoding/ascii.go` | Converts text to bytes and bytes back to text. |
| `internal/encoding/base_converter.go` | Converts decimal byte values to and from arbitrary codex bases. |
| `internal/encoding/deserializer.go` | Deserializes encoded frames back into codex tokens. |
| `internal/encoding/encoding_test.go` | Tests encoding, decoding, serialization, and conversion behavior. |
| `internal/encoding/serializer.go` | Serializes encoded tokens into binary frame format. |
| `internal/encoding/translator.go` | High-level payload encode/decode helpers used by packet construction and processing. |

### 6.15 `internal/errors/`

| File | Purpose |
| --- | --- |
| `internal/errors/errors.go` | Shared error definitions or helpers for domain-specific failures. |

### 6.16 `internal/geometry/`

| File | Purpose |
| --- | --- |
| `internal/geometry/closest_tower_pair.go` | Finds closest tower pairs between planets for link/tower calculations. |
| `internal/geometry/coordinate_scaling.go` | Converts configured coordinates into physical scaled distances. |
| `internal/geometry/geometry_test.go` | Tests geometry calculations. |
| `internal/geometry/ring_segments.go` | Calculates ring traversal segments between towers. |
| `internal/geometry/tower_failure_test.go` | Tests tower-failure effects on geometry/ring traversal. |
| `internal/geometry/tower_placement.go` | Computes clockwise tower placement around a planet. |
| `internal/geometry/void_distance.go` | Computes interplanetary void distance between planets. |

### 6.17 `internal/intelligence/`

| File | Purpose |
| --- | --- |
| `internal/intelligence/congestion.go` | Go congestion model interface/legacy implementation helpers. |
| `internal/intelligence/history.go` | Historical observation helpers used for intelligence calculations. |
| `internal/intelligence/model_loader.go` | Loads JSON model artifacts for legacy or inspection-compatible model paths. |
| `internal/intelligence/python_service.go` | Active Go-to-Python FastAPI model client used by the runtime. |
| `internal/intelligence/targeting.go` | Targeting-risk scoring logic or legacy model support. |
| `internal/intelligence/trust.go` | Trust scoring logic or legacy model support. |
| `internal/intelligence/uncertainty.go` | Computes uncertainty scores from observations and data quality. |

### 6.18 `internal/latency/`

| File | Purpose |
| --- | --- |
| `internal/latency/atmosphere.go` | Computes atmospheric traversal latency. |
| `internal/latency/fiber.go` | Computes fiber/ring latency inside planets. |
| `internal/latency/latency_test.go` | Tests latency formulas and aggregate behavior. |
| `internal/latency/planet_transit.go` | Computes per-planet entry-to-exit transit latency. |
| `internal/latency/route_latency.go` | Aggregates planet and void transit latency into route totals. |
| `internal/latency/tower_delay.go` | Computes tower processing or transfer delay. |
| `internal/latency/tower_failure_test.go` | Tests latency behavior under tower failure. |
| `internal/latency/void.go` | Computes void-space transit latency between planets. |

### 6.19 `internal/liveapi/`

| File | Purpose |
| --- | --- |
| `internal/liveapi/client.go` | Live state provider for the agent. Currently generates validated mock state and is prepared for Chimera API integration. |
| `internal/liveapi/tick_cache.go` | Maintains current tick and recent network state cache. |
| `internal/liveapi/types.go` | Live API request/response or raw state type definitions. |
| `internal/liveapi/validator.go` | Validates live observations, link status, saturation, and usable link state. |

### 6.20 `internal/models/`

| File | Purpose |
| --- | --- |
| `internal/models/congestion_model.joblib` | Python runtime artifact for congestion prediction. |
| `internal/models/congestion_model.json` | JSON congestion model export with metadata and compatibility fields. |
| `internal/models/embed.go` | Embeds model JSON artifacts into the Go binary for legacy/inspection use. |
| `internal/models/model_manifest.json` | Consolidated metadata for trained model artifacts, metrics, features, and split strategy. |
| `internal/models/targeting_model.joblib` | Python runtime artifact for targeting-risk prediction. |
| `internal/models/targeting_model.json` | JSON targeting model export with metadata and compatibility fields. |
| `internal/models/trust_model.joblib` | Python runtime artifact for trust prediction. |
| `internal/models/trust_model.json` | JSON trust model export with metadata and compatibility fields. |

### 6.21 `internal/observability/`

| File | Purpose |
| --- | --- |
| `internal/observability/logger.go` | Shared logging helper for observability and structured messages. |

### 6.22 `internal/orchestrator/`

| File | Purpose |
| --- | --- |
| `internal/orchestrator/endpoints.go` | Resolves planet service endpoints from configuration/environment. |
| `internal/orchestrator/http_server.go` | Registers orchestrator HTTP API routes and handlers. |
| `internal/orchestrator/remote_client.go` | Client used by Wails `app.go` to call orchestrator endpoints and event streams. |
| `internal/orchestrator/service.go` | Core orchestrator service: health monitoring, graph snapshots, transmissions, node/link/tower controls, and reset. |
| `internal/orchestrator/tower_control_test.go` | Tests tower disable/enable behavior through orchestrator logic. |

### 6.23 `internal/packet/`

| File | Purpose |
| --- | --- |
| `internal/packet/factory.go` | Creates packet structs, encoded payloads, and route metadata. |
| `internal/packet/hop_log_builder.go` | Builds ordered mathematical hop logs for packet route execution. |
| `internal/packet/processor.go` | Validates decoded payload integrity. |

### 6.24 `internal/planet/`

| File | Purpose |
| --- | --- |
| `internal/planet/handler.go` | HTTP handlers for planet node health, admin state, packet receive, forwarding, and telemetry callbacks. |
| `internal/planet/metrics.go` | Planet-node metrics and status helpers. |
| `internal/planet/service.go` | Planet service logic for packet processing and forwarding along a route. |
| `internal/planet/tower_state_test.go` | Tests planet behavior when tower state changes. |

### 6.25 `internal/reliable/`

| File | Purpose |
| --- | --- |
| `internal/reliable/acknowledgement.go` | ACK data structures and acknowledgement helper logic. |
| `internal/reliable/checkpoint.go` | Tracks last confirmed packet position for safe retry/reroute behavior. |
| `internal/reliable/queue.go` | Queue for packets waiting to be sent or retried. |
| `internal/reliable/retransmission.go` | Retry/retransmission policy helpers. |
| `internal/reliable/sender.go` | Reliable packet sender abstraction with ACK timeout, retries, status transitions, and checkpoint updates. |

### 6.26 `internal/resilience/`

| File | Purpose |
| --- | --- |
| `internal/resilience/network_state.go` | Tracks node health, manual disables, disabled links, disabled towers, and reset state. |
| `internal/resilience/network_state_test.go` | Tests resilience state transitions and snapshots. |

### 6.27 `internal/routing/`

| File | Purpose |
| --- | --- |
| `internal/routing/dijkstra.go` | Expanded-state Dijkstra route search with deterministic tie-breaking and tower-failure transition skipping. |
| `internal/routing/graph.go` | Graph data structures and neighbor/link access helpers. |
| `internal/routing/graph_builder.go` | Builds routing graph from universe config, unavailable nodes, disabled links, and disabled towers. |
| `internal/routing/link_validator.go` | Validates link availability and tower-level constraints. |
| `internal/routing/priority_queue.go` | Priority queue used by Dijkstra. |
| `internal/routing/route_cost.go` | Computes route transition and destination costs. |
| `internal/routing/route_finder.go` | Builds route details and latency breakdowns from a selected path. |
| `internal/routing/routing_test.go` | Tests normal routing behavior. |
| `internal/routing/tower_failure_test.go` | Tests routing behavior when tower failures affect valid paths. |

### 6.28 `internal/simulation/`

| File | Purpose |
| --- | --- |
| `internal/simulation/event.go` | Simulation event type definitions. |
| `internal/simulation/result.go` | Simulation result type definitions. |

### 6.29 `internal/transport/`

| File | Purpose |
| --- | --- |
| `internal/transport/event_hub.go` | In-memory event pub/sub hub used for SSE streaming. |
| `internal/transport/http_client.go` | HTTP client wrapper for JSON requests between services. |
| `internal/transport/json.go` | JSON encoding/decoding helpers for HTTP transport. |

### 6.30 `pkg/protocol/`

| File | Purpose |
| --- | --- |
| `pkg/protocol/event.go` | Shared event payloads for network telemetry and SSE streams. |
| `pkg/protocol/request.go` | Shared API request DTOs for transmissions, links, towers, and node actions. |
| `pkg/protocol/response.go` | Shared API response DTOs for universe and transmission calls. |

### 6.31 `tests/`

| File | Purpose |
| --- | --- |
| `tests/agent/agent_test.go` | Tests Phase 2 agent behavior. |
| `tests/agent/chaos_simulator.go` | Test helper for simulating failures or changing network conditions. |
| `tests/agent/harness_local_test.go` | Local test harness for agent integration behavior. |
| `tests/agent/harness_local_test.go.example` | Example local harness file/template. |
| `tests/agent/harness_test.go` | Shared agent test harness. |
| `tests/integration/transmission_test.go` | Integration test for end-to-end transmission behavior. |

### 6.32 `scripts/`

| File | Purpose |
| --- | --- |
| `scripts/create-project-folders.ps1` | Helper script for creating/scaffolding expected project folders. |
| `scripts/setup-windows.ps1` | Windows setup helper for development dependencies and environment preparation. |
| `scripts/verify_all.ps1` | Windows verification script for formatting, vetting, Go tests, frontend build, and Wails build. |
| `scripts/verify_all.sh` | Unix shell verification script equivalent. |

### 6.33 `patches/`

| File | Purpose |
| --- | --- |
| `patches/App.tsx.integration.snippet.txt` | Integration snippet/reference for updating the React app. |
| `patches/main.go.binding.snippet.txt` | Integration snippet/reference for Wails `main.go` bindings. |
| `patches/styles.css.integration.snippet.txt` | Integration snippet/reference for CSS changes. |

### 6.34 `build/`

| File | Purpose |
| --- | --- |
| `build/appicon.png` | Application icon source image. |
| `build/windows/icon.ico` | Windows icon for Wails builds. |
| `build/windows/info.json` | Windows build metadata. |
| `build/windows/wails.exe.manifest` | Windows executable manifest used by Wails. |

### 6.35 Hidden Project Support Files

| File | Purpose |
| --- | --- |
| `.github/workflows/ci.yml` | GitHub Actions CI workflow for automated checks. |
| `.vscode/extensions.json` | Recommended VS Code extensions for contributors. |
| `.vscode/tasks.json` | VS Code task definitions for common project commands. |

### 6.36 Local-Only Directories Seen in This Workspace

These directories may exist on a developer machine but are not core implementation files:

| Directory | Purpose |
| --- | --- |
| `.git/` | Local Git repository metadata. Do not edit manually. |
| `.agents/` | Local agent/tooling workspace metadata. Not part of the runtime application. |
| `.pytest_cache/` | Pytest cache generated after running Python tests. |
| `.venv/` | Local Python virtual environment. Recreate from `training/requirements.txt`. |

## 7. Important Generated, Binary, and Artifact Files

Some files should not be treated like hand-authored source:

- `frontend/wailsjs/**` is generated by Wails.
- `frontend/package-lock.json` is generated by npm but should be committed for reproducible installs.
- `internal/models/*.joblib` are binary Python model artifacts.
- `internal/models/*.json` and `internal/models/model_manifest.json` are generated model exports/metadata.
- `training/reports/*` are generated reports.
- `frontend/public/planets/*.webp` are UI image assets.
- `build/windows/*` are Wails/Windows build support files.

## 8. Key Talking Points for a Presentation

- The project is not only a UI. The route engine, physics calculations, packet processing, failure state, and ML decision logic are implemented in backend code.
- The frontend does not compute official routing or latency formulas. It calls Wails bindings and displays backend results.
- The ML workflow prevents time leakage by splitting by tick before feature engineering and using shifted rolling features.
- The runtime uses a FastAPI model service so Go can consume the best sklearn `.joblib` models without reimplementing sklearn.
- The agent decision is explainable because each decision stores physical latency, congestion penalty, trust score, targeting-risk score, uncertainty, combined cost, and reasons.
- Failure handling is graph-based: disabled nodes, links, and towers are removed from routing decisions before a route is selected.
- Audit and timeline records are kept for judge-facing traceability.

## 9. Recommended Demo Sequence

1. Start the ML service.
2. Start the Docker network.
3. Start the Wails app.
4. Show the universe map and baseline route.
5. Send a simple packet and show hop logs/latency.
6. Disable a node or tower and show that the next route changes.
7. Open the agent panel and evaluate a transmission request.
8. Explain the link evaluation table using true-cost components.
9. Show the decision audit log as evidence of explainability.
10. Show `training/reports/model_selection_summary.csv` and `internal/models/model_manifest.json` for ML provenance.
