package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// InferenceClient is an HTTP client that calls the Python FastAPI
// inference server running alongside the Go process.
//
// The server exposes a single endpoint:
//
//	POST /predict   →   accepts InferenceRequest, returns InferenceResponse
//
// This design avoids subprocess overhead and gives us clean JSON contracts
// between Go and Python.
type InferenceClient struct {
	baseURL string
	client  *http.Client
}

// NewInferenceClient creates a client pointing at the Python inference server.
// URL is read from the AI_SERVER_URL env variable; defaults to localhost:7070.
func NewInferenceClient() *InferenceClient {
	base := strings.TrimRight(os.Getenv("AI_SERVER_URL"), "/")
	if base == "" {
		base = "http://localhost:7070"
	}
	return &InferenceClient{
		baseURL: base,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// Predict sends a single link's live state to the Python server and
// receives all three model scores in one round-trip.
func (c *InferenceClient) Predict(ctx context.Context, req InferenceRequest) (InferenceResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return InferenceResponse{}, fmt.Errorf("inference: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/predict",
		bytes.NewReader(body),
	)
	if err != nil {
		return InferenceResponse{}, fmt.Errorf("inference: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return InferenceResponse{}, fmt.Errorf("inference: POST /predict: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return InferenceResponse{}, fmt.Errorf(
			"inference: POST /predict returned HTTP %d for link %s",
			resp.StatusCode, req.LinkID,
		)
	}

	var out InferenceResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return InferenceResponse{}, fmt.Errorf("inference: decode response: %w", err)
	}
	return out, nil
}

// Ping checks whether the inference server is reachable.
// Returns nil if the server responds with HTTP 200 on GET /health.
func (c *InferenceClient) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("inference server unreachable at %s: %w", c.baseURL, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("inference server /health returned HTTP %d", resp.StatusCode)
	}
	return nil
}
