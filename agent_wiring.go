package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	agentpkg "github.com/launch26/relic-ring-protocol/internal/agent"
	"github.com/launch26/relic-ring-protocol/internal/audit"
	"github.com/launch26/relic-ring-protocol/internal/candidate"
	"github.com/launch26/relic-ring-protocol/internal/intelligence"
	"github.com/launch26/relic-ring-protocol/internal/liveapi"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// buildAgentBindings constructs the real internal agent and wraps it in the
// DTO-based AgentBindings expected by agent_app.go.
//
// ------------------------------------------------------------------
// LIVE API KEY INTEGRATION:
//   Set the following environment variables before launching the app:
//     CHIMERA_API_KEY=<your-api-key>
//     CHIMERA_BASE_URL=https://chimera.launch26.space
//   Without CHIMERA_API_KEY the mock state provider will be used.
// ------------------------------------------------------------------
func buildAgentBindings() (AgentBindings, error) {
	cfg, err := agentpkg.LoadConfig("configs/agent_config.json")
	if err != nil {
		cfg = agentpkg.DefaultConfig()
	}

	planets, links, baselines, capacities := loadTopology()

	modelsDir := "internal/models"
	congestion, err := intelligence.NewCongestionPredictor(modelsDir + "/congestion_model.json")
	if err != nil {
		log.Printf("WARNING: failed to load congestion model: %v (predictions will be zero)", err)
	}
	trust, err := intelligence.NewTrustScorer(modelsDir + "/trust_model.json")
	if err != nil {
		log.Printf("WARNING: failed to load trust model: %v (predictions will be zero)", err)
	}
	targeting, err := intelligence.NewTargetingScorer(modelsDir + "/targeting_model.json")
	if err != nil {
		log.Printf("WARNING: failed to load targeting model: %v (predictions will be zero)", err)
	}

	apiKey := os.Getenv("CHIMERA_API_KEY")
	baseURL := os.Getenv("CHIMERA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://chimera.launch26.space"
	}
	liveClient := liveapi.NewClient(baseURL, apiKey, baselines, capacities)

	var rLinks []agentpkg.RouterLink
	var cLinks []candidate.LinkDef
	for _, l := range links {
		rLinks = append(rLinks, agentpkg.RouterLink{A: l.from, B: l.to, LatencyMS: l.latencyMS})
		cLinks = append(cLinks, candidate.LinkDef{A: l.from, B: l.to, LatencyMS: l.latencyMS})
	}

	router := agentpkg.NewRouterAdapter(planets, rLinks)
	gen := candidate.NewGenerator(planets, cLinks)
	loop := agentpkg.NewDecisionLoop(cfg, congestion, trust, targeting, liveClient)
	parser := agentpkg.NewParser(planets)

	ag := &agentpkg.Agent{
		Config:     cfg,
		Parser:     parser,
		Router:     router,
		Candidates: gen,
		Congestion: congestion,
		Trust:      trust,
		Targeting:  targeting,
		LiveState:  liveClient,
		Loop:       loop,
	}

	// Simple timeline tracker.
	var timelineMu sync.Mutex
	var timeline []PacketTimelineEntryDTO

	// Define the onHop callback to stream execution.
	ag.OnHop = func(ctx context.Context, hop agentpkg.HopDecision) {
		runtime.EventsEmit(ctx, "agent:hop", hop)
	}
	ag.Loop.OnHop = ag.OnHop

	auditLog := audit.NewLogger(500)

	bindings := AgentBindings{
		Parse: func(raw string) (ParsedTransmissionRequestDTO, error) {
			parsed, err := ag.Parser.Parse(raw)
			if err != nil {
				return ParsedTransmissionRequestDTO{}, err
			}
			return ParsedTransmissionRequestDTO{
				OriginID:      parsed.OriginID,
				DestinationID: parsed.DestinationID,
				Payload:       parsed.Payload,
				Confidence:    1.0,
				Ambiguities:   []string{},
			}, nil
		},

		Execute: func(ctx context.Context, raw string) (DecisionReportDTO, error) {
			report, err := ag.Execute(ctx, raw)
			if err != nil {
				return DecisionReportDTO{}, err
			}

			// Log hop decisions to audit.
			for i, hop := range report.HopDecisions {
				auditLog.Log(audit.NewRecord(
					hop.Tick, hop.CurrentPlanet, hop.LinkID,
					string(hop.Action),
					hop.Evaluation.PhysicalLatencyMS,
					hop.Evaluation.PredictedCongestionPenaltyMS,
					hop.Evaluation.TrustScore,
					hop.Evaluation.TargetingRiskScore,
					hop.Evaluation.UncertaintyScore,
					hop.Evaluation.CombinedCost,
					hop.Reasons, hop.NextPlanet, hop.AlternativePath,
					len(report.LinkEvaluations), 0.85,
				))

				// Track packet timeline entries.
				timelineMu.Lock()
				timeline = append(timeline, PacketTimelineEntryDTO{
					MessageID:      fmt.Sprintf("%s->%s", report.OriginID, report.DestinationID),
					PacketID:       fmt.Sprintf("pkt-%d", i+1),
					SequenceNumber: i + 1,
					TotalPackets:   len(report.HopDecisions),
					State:          string(hop.Action),
					PlanetID:       hop.CurrentPlanet,
					LinkID:         hop.LinkID,
					RouteVersion:   1,
					RetryCount:     0,
					Tick:           hop.Tick,
					Detail:         fmt.Sprintf("→ %s", hop.NextPlanet),
				})
				timelineMu.Unlock()
			}

			evals := make([]LinkEvaluationDTO, 0, len(report.LinkEvaluations))
			for _, e := range report.LinkEvaluations {
				evals = append(evals, LinkEvaluationDTO{
					LinkID:                       e.LinkID,
					PhysicalLatencyMS:            e.PhysicalLatencyMS,
					PredictedCongestionPenaltyMS: e.PredictedCongestionPenaltyMS,
					TrustScore:                   e.TrustScore,
					TargetingRiskScore:           e.TargetingRiskScore,
					UncertaintyScore:             e.UncertaintyScore,
					CombinedCost:                 e.CombinedCost,
					Reasons:                      nil,
				})
			}

			// Estimate total latency as sum of physical + congestion on chosen path.
			totalMS := 0.0
			for _, e := range evals {
				totalMS += e.PhysicalLatencyMS + e.PredictedCongestionPenaltyMS
			}
			if totalMS == 0 || math.IsNaN(totalMS) {
				totalMS = report.FinalLatencyEstimate
			}

			return DecisionReportDTO{
				OriginID:               report.OriginID,
				DestinationID:          report.DestinationID,
				ChosenPath:             report.ChosenPath,
				LinkEvaluations:        evals,
				FinalLatencyEstimateMS: totalMS,
				Explanation:            report.Explanation,
				Status:                 "DELIVERED",
			}, nil
		},

		State: func() AgentStateDTO {
			return AgentStateDTO{
				Status:       "READY",
				CurrentTick:  time.Now().Unix() % 10000,
				QueuedPackets: 0,
			}
		},

		Audit: func() []AuditRecordDTO {
			records := auditLog.All()
			dtos := make([]AuditRecordDTO, 0, len(records))
			for i, r := range records {
				dtos = append(dtos, AuditRecordDTO{
					AuditID:                      fmt.Sprintf("audit-%d", i+1),
					Tick:                         r.Tick,
					CurrentPlanet:                r.CurrentPlanet,
					LinkID:                       r.LinkID,
					Action:                       r.Action,
					PhysicalLatencyMS:            r.PhysicalLatencyMS,
					PredictedCongestionPenaltyMS: r.PredictedCongestionMS,
					TrustScore:                   r.TrustScore,
					TargetingRiskScore:           r.TargetingRiskScore,
					UncertaintyScore:             r.UncertaintyScore,
					CombinedCost:                 r.CombinedCost,
					Reasons:                      r.Reasons,
					AlternativePath:              r.AlternativePath,
					Timestamp:                    r.Timestamp.Format(time.RFC3339),
				})
			}
			return dtos
		},

		Timeline: func() []PacketTimelineEntryDTO {
			timelineMu.Lock()
			defer timelineMu.Unlock()
			result := make([]PacketTimelineEntryDTO, len(timeline))
			copy(result, timeline)
			return result
		},

		Reset: func() error {
			auditLog.Clear()
			timelineMu.Lock()
			timeline = nil
			timelineMu.Unlock()
			return nil
		},
	}

	return bindings, nil
}

// linkTopo is a private helper for topology loading.
type linkTopo struct{ from, to string; latencyMS float64 }

// loadTopology reads the universe config and computes physical latencies.
func loadTopology() (planets []string, links []linkTopo, baselines, capacities map[string]float64) {
	baselines = make(map[string]float64)
	capacities = make(map[string]float64)

	data, err := os.ReadFile("datasets/universe-config.json")
	if err != nil {
		data, err = os.ReadFile("configs/universe-config.json")
		if err != nil {
			planets = []string{"Aegis", "Boreas", "Dawn", "Elysium", "Fenix", "Caelum"}
			return
		}
	}

	var cfg struct {
		Metadata struct {
			SpeedOfLightKMS float64 `json:"speed_of_light_kms"`
			ScaleUnitKM     float64 `json:"coordinate_scale_unit_km"`
		} `json:"universe_metadata"`
		Nodes []struct {
			ID string  `json:"id"`
			X  float64 `json:"x"`
			Y  float64 `json:"y"`
		} `json:"nodes"`
		Links []struct {
			LinkID        string  `json:"link_id"`
			PlanetA       string  `json:"planet_a"`
			PlanetB       string  `json:"planet_b"`
			CapacityUnits float64 `json:"capacity_units"`
		} `json:"interplanetary_links"`
	}

	if json.Unmarshal(data, &cfg) != nil {
		planets = []string{"Aegis", "Boreas", "Dawn", "Elysium", "Fenix", "Caelum"}
		return
	}

	nodeMap := make(map[string]struct{ X, Y float64 })
	for _, n := range cfg.Nodes {
		planets = append(planets, n.ID)
		nodeMap[n.ID] = struct{ X, Y float64 }{n.X, n.Y}
	}

	sol := cfg.Metadata.SpeedOfLightKMS
	scale := cfg.Metadata.ScaleUnitKM
	if sol == 0 {
		sol = 299792.0
	}
	if scale == 0 {
		scale = 1e9
	}

	for _, l := range cfg.Links {
		a, b := nodeMap[l.PlanetA], nodeMap[l.PlanetB]
		dx := (a.X - b.X) * scale
		dy := (a.Y - b.Y) * scale
		dist := math.Sqrt(dx*dx + dy*dy)
		latencyMS := (dist / sol) * 1000.0

		links = append(links, linkTopo{from: l.PlanetA, to: l.PlanetB, latencyMS: latencyMS})
		baselines[l.LinkID] = latencyMS
		capacities[l.LinkID] = l.CapacityUnits
	}
	return
}
