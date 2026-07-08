package agenttest

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"
)

func TestSaturatedLinkIsNeverSelected(t *testing.T) {
	harness := newHarness(t)
	scenario := NewScenario("saturated-primary", baselineSnapshot()).
		AtTick(2, Saturate("Aegis-Elysium")).
		UntilTick(4).
		Build()

	report, err := runWithTimeout(
		t,
		harness,
		`Send "HELLO" from Aegis to Caelum`,
		scenario,
	)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	for _, link := range report.EvaluatedLinks {
		if link.LinkID == "Aegis-Elysium" && link.Selected {
			t.Fatal("saturated link was selected")
		}
	}
}

func TestMidFlightFailureReroutesWithoutResendingConfirmedPackets(
	t *testing.T,
) {
	harness := newHarness(t)
	scenario := NewScenario("mid-flight-sever", baselineSnapshot()).
		AtTick(3, Jam("Dawn-Fenix")).
		UntilTick(6).
		Build()

	report, err := runWithTimeout(
		t,
		harness,
		`Send "PACKET CHECK" from Aegis to Caelum`,
		scenario,
	)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if report.ResentConfirmed {
		t.Fatal("agent resent a packet that had already been confirmed")
	}
	if report.AuditCount == 0 {
		t.Fatal("expected a decision audit")
	}
}

func TestNetworkPartitionQueuesSafely(t *testing.T) {
	harness := newHarness(t)
	scenario := NewScenario("partition", baselineSnapshot()).
		AtTick(
			2,
			Saturate("Aegis-Boreas"),
			Saturate("Aegis-Dawn"),
			Saturate("Aegis-Elysium"),
		).
		UntilTick(5).
		Build()

	report, err := runWithTimeout(
		t,
		harness,
		`Send "WAIT SAFELY" from Aegis to Caelum`,
		scenario,
	)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if report.Status != "NETWORK_PARTITION" &&
		report.Status != "QUEUED" {
		t.Fatalf("expected queued/partition status, got %q", report.Status)
	}
	if report.QueuedPackets == 0 {
		t.Fatal("expected at least one queued packet")
	}
}

func TestSpoofedTelemetryLowersTrust(t *testing.T) {
	harness := newHarness(t)
	scenario := NewScenario("telemetry-spoof", baselineSnapshot()).
		AtTick(2, SpoofLatency("Aegis-Elysium", 1.0)).
		UntilTick(4).
		Build()

	report, err := runWithTimeout(
		t,
		harness,
		`Send "TRUST TEST" from Aegis to Caelum`,
		scenario,
	)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	found := false
	for _, link := range report.EvaluatedLinks {
		if link.LinkID != "Aegis-Elysium" {
			continue
		}
		found = true
		if link.TrustScore >= 0.8 {
			t.Fatalf(
				"expected suspicious trust score, got %.3f",
				link.TrustScore,
			)
		}
	}
	if !found {
		t.Fatal("spoofed link was not evaluated")
	}
}

func TestAllScoresAreFiniteAndBounded(t *testing.T) {
	harness := newHarness(t)
	report, err := runWithTimeout(
		t,
		harness,
		`Send "SCORE CHECK" from Boreas to Caelum`,
		NewScenario("normal", baselineSnapshot()).
			UntilTick(3).
			Build(),
	)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	for _, link := range report.EvaluatedLinks {
		assertUnitScore(t, link.LinkID, "trust", link.TrustScore)
		assertUnitScore(
			t,
			link.LinkID,
			"targeting",
			link.TargetingRiskScore,
		)
		assertUnitScore(
			t,
			link.LinkID,
			"uncertainty",
			link.UncertaintyScore,
		)
		if math.IsNaN(link.CombinedCost) ||
			math.IsInf(link.CombinedCost, 0) ||
			link.CombinedCost < 0 {
			t.Fatalf(
				"%s has invalid combined cost %v",
				link.LinkID,
				link.CombinedCost,
			)
		}
	}
}

func TestChosenPathMatchesEndpoints(t *testing.T) {
	harness := newHarness(t)
	report, err := runWithTimeout(
		t,
		harness,
		`Send "PATH CHECK" from Aegis to Caelum`,
		NewScenario("normal", baselineSnapshot()).
			UntilTick(3).
			Build(),
	)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(report.ChosenPath) < 2 {
		t.Fatalf("chosen path is incomplete: %#v", report.ChosenPath)
	}
	if report.ChosenPath[0] != "Aegis" {
		t.Fatalf("path starts at %q", report.ChosenPath[0])
	}
	if report.ChosenPath[len(report.ChosenPath)-1] != "Caelum" {
		t.Logf("Explanation: %s", report.Explanation)
		t.Logf("Report: %+v", report)
		t.Fatalf(
			"path ends at %q",
			report.ChosenPath[len(report.ChosenPath)-1],
		)
	}
	if strings.TrimSpace(report.Explanation) == "" {
		t.Fatal("decision explanation is empty")
	}
}

func runWithTimeout(
	t *testing.T,
	harness Harness,
	request string,
	scenario Scenario,
) (DecisionReport, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()
	return harness.Run(ctx, request, scenario)
}

func assertUnitScore(
	t *testing.T,
	linkID string,
	name string,
	value float64,
) {
	t.Helper()
	if math.IsNaN(value) ||
		math.IsInf(value, 0) ||
		value < 0 ||
		value > 1 {
		t.Fatalf(
			"%s has invalid %s score %v",
			linkID,
			name,
			value,
		)
	}
}

func baselineSnapshot() Snapshot {
	latency := 35.0
	return Snapshot{
		Tick: 1,
		Links: map[string]LinkState{
			"Aegis-Boreas": {
				LinkID: "Aegis-Boreas",
				PlanetA: "Aegis",
				PlanetB: "Boreas",
				CapacityUnits: 208,
				LoadRatio: 0.35,
				CurrentLoad: 72.8,
				SelfReportedLatencyMS: &latency,
				TrafficShare: 0.08,
				Status: "ok",
			},
			"Aegis-Dawn": {
				LinkID: "Aegis-Dawn",
				PlanetA: "Aegis",
				PlanetB: "Dawn",
				CapacityUnits: 98,
				LoadRatio: 0.42,
				CurrentLoad: 41.16,
				SelfReportedLatencyMS: &latency,
				TrafficShare: 0.10,
				Status: "ok",
			},
			"Aegis-Elysium": {
				LinkID: "Aegis-Elysium",
				PlanetA: "Aegis",
				PlanetB: "Elysium",
				CapacityUnits: 202,
				LoadRatio: 0.38,
				CurrentLoad: 76.76,
				SelfReportedLatencyMS: &latency,
				TrafficShare: 0.11,
				Status: "ok",
			},
			"Boreas-Fenix": {
				LinkID: "Boreas-Fenix",
				PlanetA: "Boreas",
				PlanetB: "Fenix",
				CapacityUnits: 169,
				LoadRatio: 0.33,
				CurrentLoad: 55.77,
				SelfReportedLatencyMS: &latency,
				TrafficShare: 0.07,
				Status: "ok",
			},
			"Dawn-Fenix": {
				LinkID: "Dawn-Fenix",
				PlanetA: "Dawn",
				PlanetB: "Fenix",
				CapacityUnits: 157,
				LoadRatio: 0.30,
				CurrentLoad: 47.1,
				SelfReportedLatencyMS: &latency,
				TrafficShare: 0.08,
				Status: "ok",
			},
			"Caelum-Dawn": {
				LinkID: "Caelum-Dawn",
				PlanetA: "Caelum",
				PlanetB: "Dawn",
				CapacityUnits: 92,
				LoadRatio: 0.37,
				CurrentLoad: 34.04,
				SelfReportedLatencyMS: &latency,
				TrafficShare: 0.09,
				Status: "ok",
			},
			"Caelum-Elysium": {
				LinkID: "Caelum-Elysium",
				PlanetA: "Caelum",
				PlanetB: "Elysium",
				CapacityUnits: 83,
				LoadRatio: 0.34,
				CurrentLoad: 28.22,
				SelfReportedLatencyMS: &latency,
				TrafficShare: 0.08,
				Status: "ok",
			},
			"Caelum-Fenix": {
				LinkID: "Caelum-Fenix",
				PlanetA: "Caelum",
				PlanetB: "Fenix",
				CapacityUnits: 127,
				LoadRatio: 0.31,
				CurrentLoad: 39.37,
				SelfReportedLatencyMS: &latency,
				TrafficShare: 0.07,
				Status: "ok",
			},
		},
	}
}
