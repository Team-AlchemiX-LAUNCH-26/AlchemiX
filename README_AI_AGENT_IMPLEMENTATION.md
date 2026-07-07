# AlchemiX Phase 2 — AI Agent Implementation Guide

This document explains how to build the **Chimera-Resilient Analytical Co-Pilot Agent** for the AlchemiX Phase 2 challenge.

The agent extends the existing Phase 1 physics-based router rather than replacing it. Its role is to evaluate live interplanetary links, predict congestion, estimate telemetry trust, measure targeting risk, calculate a dynamic **True Cost**, and decide whether to keep, override, pause, reroute, or safely queue packet transmissions.

---

## 1. Can the agent be built from the available files?

Yes.

The provided materials are sufficient to build the complete agent core from scratch, including:

- Dataset loaders
- Model-training scripts
- Congestion prediction
- Telemetry trust scoring
- Targeting-risk prediction
- Live Chimera API integration
- Sequential route evaluation
- True Cost calculation
- Packet checkpointing
- Selective retransmission
- Decision auditing
- Chaos simulation and tests

However, the currently available files contain the project handoff, universe configuration, challenge brief, resilience design, and datasets—not the full Go/Wails repository.

This means the agent core can be built independently, but clean integration into the existing Phase 1 application requires the real backend packages, especially:

```text
go.mod
app.go
main.go
internal/routing/*
internal/latency/*
internal/packet/*
internal/resilience/*
internal/service/*
```

The recommended approach is therefore:

1. Build the agent as a separate internal module.
2. Define interfaces around the existing Phase 1 router.
3. Integrate the real router once the source repository is available.

---

## 2. What the AI agent should actually be

The agent should not be a chatbot that freely invents routes.

It should be an **explainable decision engine** that coordinates deterministic tools and trained analytical models.

```text
Natural-language request
          ↓
Request Parser
          ↓
Phase 1 Physics Router
          ↓
Top-K Candidate Routes
          ↓
Sequential Co-Pilot Agent
    ├── Congestion Model
    ├── Telemetry Trust Model
    ├── Targeting-Risk Model
    ├── Uncertainty Detector
    └── True Cost Calculator
          ↓
Safe next-hop decision
          ↓
Reliable packet executor
          ↓
ACK / checkpoint / reroute
          ↓
Standardized decision report
```

An LLM may optionally help parse flexible user input, but the actual routing decision should remain:

- Numerical
- Reproducible
- Testable
- Auditable
- Configurable

---

## 3. Core responsibilities of the agent

The Phase 2 agent must perform the following steps:

1. Parse the origin, destination, and payload from natural-language input.
2. Generate the Phase 1 physics-based baseline route.
3. Generate alternative candidate routes.
4. Fetch the latest live network state.
5. Validate the state and remove unusable links.
6. Predict congestion penalties.
7. Estimate telemetry trust.
8. Estimate Chimera targeting risk.
9. Calculate the True Cost of every candidate link.
10. Select the safest next hop.
11. Transmit using packet checkpointing.
12. Reroute only unconfirmed packets when a link fails.
13. Queue safely when no valid route exists.
14. Produce an explainable decision audit.

---

## 4. Natural-language request parser

The user may enter a request such as:

```text
Send "Emergency supplies ready" from Aegis to Caelum.
```

The parser should produce:

```json
{
  "origin_id": "Aegis",
  "destination_id": "Caelum",
  "payload": "Emergency supplies ready"
}
```

### Recommended parsing strategy

Because the universe contains a small known set of planets, a deterministic parser is safer than relying entirely on an LLM.

The parser should:

1. Load valid planet names dynamically from `universe-config.json`.
2. Normalize capitalization and punctuation.
3. Detect origin keywords such as `from`.
4. Detect destination keywords such as `to`.
5. Extract quoted text as the payload where possible.
6. Validate that origin and destination are different.
7. Reject ambiguous input instead of guessing.

### Parser interface

```go
type ParsedTransmissionRequest struct {
    OriginID      string
    DestinationID string
    Payload       string
}

type RequestParser interface {
    Parse(raw string) (ParsedTransmissionRequest, error)
}
```

---

## 5. Phase 1 baseline router

The Phase 1 route-finder remains the source of truth for physical validity.

Its responsibilities are:

```text
Origin + destination
        ↓
Build active graph
        ↓
Remove failed planets and links
        ↓
Enforce maximum void-hop distance
        ↓
Calculate physical latency
        ↓
Return lowest-latency valid path
```

The Phase 2 agent must wrap this router rather than replace it.

### Router interface

```go
type Route struct {
    Path             []string
    PhysicalLatencyMS float64
}

type BaselineRouter interface {
    FindRoute(originID, destinationID string) (Route, error)
}
```

---

## 6. Candidate route generation

Using only one shortest path makes the system predictable.

The agent should generate several valid paths, for example:

```text
Candidate 1: Aegis → Elysium → Caelum
Candidate 2: Aegis → Dawn → Caelum
Candidate 3: Aegis → Boreas → Fenix → Caelum
Candidate 4: Aegis → Dawn → Fenix → Caelum
Candidate 5: Aegis → Boreas → Elysium → Caelum
```

A practical implementation can use **Yen's K-shortest paths algorithm**.

### Candidate ranking factors

Each route should be evaluated using:

- Physical latency
- Number of hops
- Congestion risk
- Telemetry trust
- Targeting risk
- Recent route usage
- Shared-edge exposure
- Uncertainty

A backup route should be edge-diverse. A second route that shares most of the same links is not a strong backup during coordinated multi-link jamming.

### Candidate generator interface

```go
type CandidateGenerator interface {
    TopKPaths(
        originID string,
        destinationID string,
        k int,
        excludedLinks map[string]bool,
    ) ([]Route, error)
}
```

---

# 7. Analytical models

The agent should use three specialized analytical models.

---

## 7.1 Congestion prediction model

### Training data

Use `link_traffic_history.csv`.

Important fields:

```text
link_id
 tick
load_units
load_ratio
status
observed_latency_ms
```

### Required predictions

For each link, the congestion model should produce:

```json
{
  "predicted_congestion_penalty_ms": 18.4,
  "saturation_probability": 0.07,
  "confidence": 0.88
}
```

### Recommended model design

Use two related predictors:

1. **Saturation classifier**

```text
P(status == saturated)
```

2. **Congestion penalty regressor**

```text
observed latency - physical baseline latency
```

### Suggested features

```text
load_ratio
current_load
capacity_units
link_id
physical_latency_ms
previous_load_ratio
load_ratio_change
rolling_average_load
rate_of_load_increase
distance_to_saturation
```

### Hard safety rule

```text
load_ratio >= 0.90 → unavailable
```

A saturated link must be removed from the active graph. A null latency must never be interpreted as zero latency.

### Congestion model interface

```go
type CongestionPrediction struct {
    PenaltyMS             float64
    SaturationProbability float64
    Confidence            float64
}

type CongestionModel interface {
    Predict(observation LinkObservation) CongestionPrediction
}
```

---

## 7.2 Telemetry trust model

### Training data

Use `link_telemetry.csv`.

Important fields:

```text
link_id
 tick
self_reported_latency_ms
measured_latency_ms
```

### Training target

Calculate the under-reporting ratio:

```text
under_report_ratio =
    max(0, measured_latency - self_reported_latency)
    / measured_latency
```

### Live inference problem

During live operation, measured latency is hidden.

The trust model must compare the self-reported value with:

- Physical baseline latency
- Predicted congestion penalty
- Historical link bias
- Previous self-reported values
- Plausible physical minimum
- Expected latency distribution

### Example

```text
Physical latency                 = 32 ms
Predicted congestion penalty     = 15 ms
Predicted realistic latency      = 47 ms
Self-reported latency            = 29 ms
Estimated under-reporting        ≈ 38%
Result                           = low trust
```

### Trust score

```text
trust_score ∈ [0, 1]
```

Suggested interpretation:

| Trust score | Meaning |
|---:|---|
| 0.80–1.00 | Trusted |
| 0.60–0.79 | Cautious |
| 0.40–0.59 | Suspicious |
| 0.00–0.39 | Quarantine candidate |

These thresholds should be configurable.

### Trust model interface

```go
type ScoreResult struct {
    Score      float64
    Confidence float64
    Reasons    []string
}

type TrustModel interface {
    Score(
        observation LinkObservation,
        congestion CongestionPrediction,
    ) ScoreResult
}
```

---

## 7.3 Targeting-risk model

### Training data

Use `link_incident_history.csv`.

Important fields:

```text
link_id
 tick
traffic_share
jammed_flag
```

### Required prediction

```text
P(link will be jammed during the current decision window)
```

### Suggested features

```text
traffic_share
traffic_share_rank
link_historical_jam_rate
recent_route_usage
consecutive_usage_count
selection_frequency
recent_load_increase
link_id
```

### Two-part risk calculation

```text
targeting risk =
    historical jam probability
    + self-created predictability penalty
```

Example:

```text
Historical jam probability       0.10
High traffic-share effect        0.08
Used in last four transmissions  0.12
Combined targeting risk          0.30
```

### Targeting model interface

```go
type UsageHistory struct {
    RecentSelections  map[string]int
    ConsecutiveUses   map[string]int
    LastSelectedTick  map[string]int64
}

type TargetingModel interface {
    Score(
        observation LinkObservation,
        usage UsageHistory,
    ) ScoreResult
}
```

---

## 7.4 Uncertainty detector

The live event may contain conditions not represented in the training data.

The uncertainty detector should flag:

- Missing fields
- Invalid ranges
- Unknown statuses
- Contradictory values
- Stale ticks
- Extreme out-of-distribution values
- Unexpected null values
- Impossible latency values

### Uncertainty score

```text
uncertainty_score ∈ [0, 1]
```

High uncertainty should increase the True Cost and may trigger:

- Probe-before-commit
- Smaller packet windows
- Conservative fallback
- Queueing
- Manual warning

---

# 8. True Cost calculation

For each usable link:

```text
True Cost =
    physical latency
  + predicted congestion penalty
  + trust penalty
  + targeting-risk penalty
  + uncertainty penalty
  + switching penalty
```

### Penalty equations

```text
trust_penalty =
    trust_weight_ms × (1 - trust_score)

risk_penalty =
    targeting_weight_ms × targeting_risk_score

uncertainty_penalty =
    uncertainty_weight_ms × uncertainty_score
```

### Example

```text
Physical latency                  31.8 ms
Congestion penalty                18.4 ms
Trust score                       0.37
Trust weight                      40.0 ms
Targeting risk                    0.62
Targeting weight                  25.0 ms
Uncertainty                       0.14
Uncertainty weight                20.0 ms
```

Calculations:

```text
Trust penalty = 40 × (1 - 0.37)
              = 25.2 ms

Targeting penalty = 25 × 0.62
                  = 15.5 ms

Uncertainty penalty = 20 × 0.14
                    = 2.8 ms

True Cost = 31.8 + 18.4 + 25.2 + 15.5 + 2.8
          = 93.7 ms
```

### Configuration

Weights must not be hidden constants.

```json
{
  "trust_weight_ms": 40.0,
  "targeting_weight_ms": 25.0,
  "uncertainty_weight_ms": 20.0,
  "switching_weight_ms": 10.0,
  "saturation_load_ratio": 0.90,
  "top_k_routes": 5
}
```

---

# 9. Sequential decision loop

The co-pilot must decide one hop at a time.

```go
for currentPlanet != destination {
    state := liveClient.GetLatestState()

    validatedState := validator.Validate(state)

    candidates := routeGenerator.TopKPaths(
        currentPlanet,
        destination,
        validatedState,
    )

    scoredCandidates := scorer.Score(
        candidates,
        validatedState,
    )

    nextHop := policy.SelectNextHop(scoredCandidates)

    result := transmitter.SendPacketWindow(
        currentPlanet,
        nextHop,
        pendingPackets,
    )

    if result.Acknowledged {
        checkpoint.AdvanceTo(nextHop)
        currentPlanet = nextHop
        continue
    }

    quarantine.MarkUncertain(result.LinkID)
    routeGenerator.Exclude(result.LinkID)
}
```

The agent does not commit blindly to the complete route.

It commits only to:

```text
Current planet → next safe planet
```

Then it refreshes the live state and evaluates again.

---

# 10. Reliable packet transmission

The analytical models only decide which link appears safest.

Zero-loss recovery requires a packet-reliability layer.

```text
Message
  ↓
Split into packets
  ↓
Store packet at current relay
  ↓
Transmit packet
  ↓
Wait for next-hop acknowledgement
  ↓
Release local active copy only after ACK
```

### Recommended packet structure

```go
type Packet struct {
    MessageID           string
    PacketID            string
    SequenceNumber      int
    TotalPackets        int
    Payload             []byte
    Checksum            string
    CurrentPlanet       string
    LastConfirmedPlanet string
    RouteVersion        int
    RetryCount          int
    TTL                 int
    Priority            string
    Status              PacketStatus
}
```

### Packet state values

```text
CREATED
CHUNKED
QUEUED
IN_FLIGHT
ACKNOWLEDGED
DELIVERY_UNCERTAIN
REROUTING
NETWORK_PARTITION
DELIVERED
FAILED_SAFE
EXPIRED
```

### Recovery after link failure

```text
Confirmed packets       → remain confirmed
Unconfirmed packet      → retransmit
Later packets           → remain queued
Route                   → recalculate from last confirmed planet
```

This avoids restarting the full message.

---

# 11. Pause, reroute, or queue policy

The agent has three main actions.

## Continue

Use the link when:

- Status is `ok`
- Load is below the hard threshold
- Trust is acceptable
- Risk is acceptable
- The route remains physically valid

## Reroute

Reroute immediately when:

- Link status is saturated
- Link is missing or invalid
- Probe times out
- Trust falls below the hard threshold
- ACK timeout occurs
- A significantly safer alternative exists

## Queue

Queue at the last safe relay when:

- No physically valid path remains
- Several coordinated failures disconnect the graph
- Live state is too unreliable to justify transmission
- The packet TTL has not yet expired

---

# 12. Recommended Go package structure

```text
internal/
├── agent/
│   ├── agent.go
│   ├── parser.go
│   ├── decision_loop.go
│   ├── policy.go
│   ├── true_cost.go
│   └── explanation.go
│
├── intelligence/
│   ├── congestion.go
│   ├── trust.go
│   ├── targeting.go
│   ├── uncertainty.go
│   └── model_loader.go
│
├── liveapi/
│   ├── client.go
│   ├── types.go
│   ├── validator.go
│   └── tick_cache.go
│
├── candidate/
│   ├── top_k.go
│   ├── diversity.go
│   └── recent_usage.go
│
├── reliable/
│   ├── sender.go
│   ├── acknowledgement.go
│   ├── checkpoint.go
│   ├── retransmission.go
│   └── queue.go
│
├── audit/
│   ├── record.go
│   ├── logger.go
│   └── report.go
│
└── models/
    ├── congestion_model.json
    ├── trust_model.json
    └── targeting_model.json
```

Offline training:

```text
training/
├── train_congestion.py
├── train_trust.py
├── train_targeting.py
├── validate_models.py
└── export_models.py
```

---

# 13. Core Go types and interfaces

```go
type LinkObservation struct {
    Tick                  int64
    LinkID                string
    PlanetA               string
    PlanetB               string
    CapacityUnits         float64
    CurrentLoad           float64
    LoadRatio             float64
    SelfReportedLatencyMS *float64
    TrafficShare          float64
    Status                string
    PhysicalLatencyMS     float64
}

type LinkEvaluation struct {
    LinkID                       string  `json:"link_id"`
    PredictedCongestionPenaltyMS float64 `json:"predicted_congestion_penalty_ms"`
    TrustScore                   float64 `json:"trust_score"`
    TargetingRiskScore           float64 `json:"targeting_risk_score"`
    CombinedCost                 float64 `json:"combined_cost"`
}
```

### Agent structure

```go
type Agent struct {
    Parser         RequestParser
    Router         BaselineRouter
    CandidatePaths CandidateGenerator
    Congestion     CongestionModel
    Trust          TrustModel
    Targeting      TargetingModel
    LiveState      StateProvider
    Transmitter    ReliableTransmitter
    Audit          AuditLogger
}

func (a *Agent) Execute(
    ctx context.Context,
    naturalLanguageRequest string,
) (DecisionReport, error)
```

---

# 14. Standardized public output

The competition-facing result should match the required schema exactly.

```json
{
  "origin_id": "Aegis",
  "destination_id": "Caelum",
  "chosen_path": [
    "Aegis",
    "Dawn",
    "Caelum"
  ],
  "link_evaluations": [
    {
      "link_id": "Aegis-Dawn",
      "predicted_congestion_penalty_ms": 12.4,
      "trust_score": 0.91,
      "targeting_risk_score": 0.18,
      "combined_cost": 44.2
    },
    {
      "link_id": "Caelum-Dawn",
      "predicted_congestion_penalty_ms": 9.8,
      "trust_score": 0.87,
      "targeting_risk_score": 0.22,
      "combined_cost": 41.6
    }
  ],
  "final_latency_estimate_ms": 85.8,
  "explanation": "Selected the Dawn route because it had lower targeting risk and stronger telemetry trust than the Elysium route."
}
```

### Canonical link IDs

The two planet names must always be alphabetically ordered.

```go
func CanonicalLinkID(a, b string) string {
    if a < b {
        return a + "-" + b
    }
    return b + "-" + a
}
```

Internal audit records may include more details, but the public response must preserve the exact expected fields.

---

# 15. Model-training strategy

## Avoid row-level random splitting

All links from one tick describe the same network moment.

Randomly placing rows from the same tick into both training and validation sets can cause leakage.

Use a tick-based split:

```text
Training ticks:   1–350
Validation ticks: 351–425
Testing ticks:    426–500
```

Or use grouped cross-validation:

```text
group = tick
```

## Metrics

### Congestion

```text
Mean Absolute Error
Root Mean Squared Error
Saturation precision
Saturation recall
Calibration error
```

### Trust

```text
Under-reporting prediction error
Spoofed-link precision
Spoofed-link recall
False-positive rate on honest links
Calibration error
```

### Targeting

```text
ROC-AUC
PR-AUC
Brier score
Probability calibration
Recall at selected risk threshold
```

Raw accuracy should not be the main metric because jammed events may be imbalanced.

---

# 16. Live Chimera API integration

The live client calls:

```http
GET /state
X-Team-Key: <team-key>
```

Configuration should come from environment variables:

```bash
CHIMERA_API_KEY=your-secret-key
CHIMERA_BASE_URL=https://chimera.launch26.space
```

Never hardcode the key.

### API client protections

The live client should support:

- Request timeout
- Bounded retries
- Tick deduplication
- Last-known-good snapshot
- Schema validation
- Staleness detection
- Circuit breaker
- Reasonable polling interval
- Safe handling of malformed responses

### Important training rule

Use the historical CSV files for model training.

Do not train the final models from build-period `/state` responses because those values may be intentionally scrambled.

---

# 17. Decision auditing

Every routing decision should be understandable without reading the code.

### Recommended internal audit record

```json
{
  "tick": 1042,
  "current_planet": "Elysium",
  "link_id": "Caelum-Elysium",
  "action": "REROUTE",
  "physical_latency_ms": 31.8,
  "predicted_congestion_penalty_ms": 18.4,
  "trust_score": 0.37,
  "targeting_risk_score": 0.62,
  "uncertainty_score": 0.14,
  "combined_cost": 93.7,
  "reasons": [
    "trust score below hard threshold",
    "repeated predicted under-reporting",
    "safer alternative path available"
  ],
  "alternative_path": [
    "Elysium",
    "Fenix",
    "Caelum"
  ]
}
```

The explanation should clearly separate:

- Observed facts
- Model predictions
- Engineering thresholds
- Final action
- Confidence
- Uncertainty

---

# 18. Chaos scenarios to test

The following cases should be included in the automated test suite.

## Routing and congestion

1. One link approaches saturation.
2. One link becomes saturated before transmission.
3. The currently used link fails mid-transmission.
4. Rerouting causes congestion on a backup path.
5. A link repeatedly flaps between healthy and failed.

## Telemetry deception

6. A link reports latency below the plausible physical minimum.
7. A link under-reports repeatedly across several ticks.
8. A noisy but honest link temporarily reports an unusual value.
9. A suspicious link fails a probe.

## Targeting risk

10. The most heavily used route receives high targeting risk.
11. Multiple links are jammed during the same tick.
12. The real route and decoy route are both jammed.
13. Route oscillation occurs between two similar paths.

## Packet reliability

14. An ACK arrives late.
15. A packet is duplicated.
16. Packets arrive out of order.
17. A packet checksum fails.
18. One unconfirmed packet must be retransmitted.

## API and unseen events

19. The API repeats an old tick.
20. The API response contains missing fields.
21. The API returns an unknown status.
22. No valid route exists.
23. The network recovers with a large queued backlog.
24. Packet TTL expires during a prolonged outage.

---

# 19. Recommended implementation roadmap

## Phase 1 — Data and model preparation

Build:

```text
CSV loaders
feature engineering
training scripts
validation reports
model export
```

Deliverables:

```text
training/train_congestion.py
training/train_trust.py
training/train_targeting.py
training/validate_models.py
internal/models/*.json
```

## Phase 2 — Agent core

Build:

```text
natural-language parser
live API client
model inference
True Cost engine
candidate route generator
sequential decision loop
explanations
```

## Phase 3 — Reliable transmission

Build:

```text
packet chunking
checksums
ACK handling
relay checkpoints
selective retransmission
deduplication
TTL
priority queues
```

## Phase 4 — Chaos resilience

Implement and test:

```text
single-link failure
multi-link jamming
saturation
spoofed telemetry
stale live state
malformed state
link flapping
network partition
recovery backlog
```

## Phase 5 — Wails integration

Expose Go methods such as:

```go
ParseTransmissionRequest(raw string)
EvaluateTransmission(raw string)
GetAgentState()
GetDecisionAudit()
GetPacketTimeline()
ResetAgentHistory()
```

The React frontend should display:

- Parsed request
- Baseline path
- Chosen path
- Link evaluations
- Congestion prediction
- Trust score
- Targeting risk
- Combined cost
- Packet state timeline
- Current queue state
- Human-readable explanation

---

# 20. Recommended development order

Start with the parts that are independent from the current repository:

```text
1. Inspect and clean datasets
2. Build model-training scripts
3. Export model artifacts
4. Define Go model interfaces
5. Build live API client
6. Build request parser
7. Build True Cost calculator
8. Build candidate route generator
9. Build sequential agent loop
10. Build packet reliability layer
11. Add decision audit logs
12. Add tests and simulations
13. Integrate with the existing Phase 1 router
14. Add Wails frontend bindings
```

---

# 21. Files required for final integration

To integrate the agent into the existing application without guessing, provide:

```text
go.mod
app.go
main.go

internal/routing/*
internal/latency/*
internal/packet/*
internal/resilience/*
internal/service/*
internal/domain/*

frontend/src/types/index.ts
frontend/src/App.tsx

one actual StartTransmission response
one actual live /state response
```

The most important integration inputs are:

1. The current router interface.
2. The current transmission service.
3. The packet and hop-log structs.
4. The exact `StartTransmission` response shape.

---

# 22. Final design principle

The agent should optimize for:

```text
Safe delivery
+ explainability
+ adaptability
+ packet integrity
```

—not simply minimum latency.

The final routing policy is:

```text
Parse request
→ generate Phase 1 baseline
→ fetch and validate live state
→ remove unusable links
→ generate top-K valid routes
→ score congestion, trust, targeting, and uncertainty
→ calculate True Cost
→ select one safe next hop
→ transmit with ACK checkpointing
→ reroute only unconfirmed packets
→ queue safely when no path exists
→ audit every decision
```

---

## Source files used

```text
PROJECT_HANDOFF_README.md
README_CHIMERA_RESILIENCE.md
Launch26 - Phase 02 - Challenge.pdf
universe-config.json
DATASETS.md
link_traffic_history.csv
link_telemetry.csv
link_incident_history.csv
```

