package agenttest

import (
	"context"
)

type DecisionReport struct {
	OriginID        string
	DestinationID   string
	ChosenPath      []string
	EvaluatedLinks  []EvaluatedLink
	Status          string
	Explanation     string
	QueuedPackets   int
	ResentConfirmed bool
	AuditCount      int
}

type EvaluatedLink struct {
	LinkID             string
	TrustScore         float64
	TargetingRiskScore float64
	UncertaintyScore   float64
	CombinedCost       float64
	Selected           bool
}

type Harness interface {
	Run(
		ctx context.Context,
		naturalLanguageRequest string,
		scenario Scenario,
	) (DecisionReport, error)
}

// newHarness is implemented in harness_local_test.go
