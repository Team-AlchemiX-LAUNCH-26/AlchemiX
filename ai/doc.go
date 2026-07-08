// Package ai implements the Launch26 Phase 2 AI Copilot Agent.
//
// The agent sits between the Phase 1 physics router and the final response.
// It evaluates each link in a route SEQUENTIALLY using three ML models:
//
//  1. Congestion Model  — predicts latency penalty from live load
//  2. Trust Model       — predicts node honesty (0–1 score)
//  3. Targeting Model   — predicts Chimera jam probability
//
// The agent NEVER modifies the Phase 1 routing engine.
// It only enriches the response with a TrueCost breakdown and explanation.
package ai
