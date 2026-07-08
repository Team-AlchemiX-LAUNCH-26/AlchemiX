// Package intelligence provides ML model inference for the agent.
// This file implements a HistGradientBoosting json loader and evaluator.
package intelligence

import (
	"encoding/json"
	"fmt"
	"os"
)

// HGBNode represents a single node in a flat array for a HistGradientBoosting tree.
type HGBNode struct {
	NodeID      int     `json:"node_id"`
	IsLeaf      bool    `json:"is_leaf"`
	FeatureName string  `json:"feature_name"`
	Threshold   float64 `json:"threshold"`
	LeftChild   int     `json:"left_child"`
	RightChild  int     `json:"right_child"`
	Value       float64 `json:"value"`
}

// Predict evaluates a single HGB tree (array of nodes) given a feature map.
func PredictHGBTree(nodes []HGBNode, features map[string]float64) float64 {
	if len(nodes) == 0 {
		return 0
	}
	curr := 0
	for {
		node := nodes[curr]
		if node.IsLeaf {
			return node.Value
		}
		
		val := features[node.FeatureName]
		if val <= node.Threshold {
			curr = node.LeftChild
		} else {
			curr = node.RightChild
		}
	}
}

// PredictHGBEnsemble evaluates an entire HGB model (sum of trees * learning rate).
func PredictHGBEnsemble(trees [][]HGBNode, features map[string]float64, base float64, lr float64) float64 {
	sum := 0.0
	for _, tree := range trees {
		sum += PredictHGBTree(tree, features)
	}
	return base + (lr * sum)
}

// ModelFile represents the JSON structure of an exported HGB model (v2 pipeline).
type ModelFile struct {
	ModelVersion       string             `json:"model_version"`
	Task               string             `json:"task"`
	ModelType          string             `json:"model_type"`
	Features           []string           `json:"features"`
	BasePrediction     float64            `json:"base_prediction"`
	LearningRate       float64            `json:"learning_rate"`
	HGBRTrees          [][]HGBNode        `json:"hgbr_trees"`
	
	// Trust specifics
	TrustScale         float64            `json:"trust_scale,omitempty"`
	SpoofThreshold     float64            `json:"spoof_threshold,omitempty"`
	
	// Targeting specifics
	PlattA             float64            `json:"platt_a,omitempty"`
	PlattB             float64            `json:"platt_b,omitempty"`
	
	// Mappings
	LinkIDMap           map[string]int     `json:"link_id_map"`
	PhysicalBaselinesMS map[string]float64 `json:"physical_baselines_ms,omitempty"`
	Capacities          map[string]float64 `json:"capacities,omitempty"`
}

// LoadModelFile reads and parses a JSON model file.
func LoadModelFile(path string) (*ModelFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load model %s: %w", path, err)
	}
	var m ModelFile
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse model %s: %w", path, err)
	}
	return &m, nil
}
