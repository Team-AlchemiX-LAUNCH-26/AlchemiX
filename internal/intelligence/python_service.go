package intelligence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/launch26/relic-ring-protocol/internal/agent"
)

const defaultPythonServiceURL = "http://127.0.0.1:8100"

type pythonPredictionStore struct {
	baseURL string
	client  *http.Client

	mu    sync.Mutex
	cache map[string]pythonCombinedResponse
}

// PythonCongestionClient implements agent.CongestionModel through the Python
// FastAPI service, which loads the best validation .joblib artifacts.
type PythonCongestionClient struct {
	store *pythonPredictionStore
}

// PythonTrustClient implements agent.TrustModel through the Python FastAPI service.
type PythonTrustClient struct {
	store *pythonPredictionStore
}

// PythonTargetingClient implements agent.TargetingModel through the Python FastAPI service.
type PythonTargetingClient struct {
	store *pythonPredictionStore
}

type pythonCombinedResponse struct {
	LinkID        string                   `json:"link_id"`
	Congestion    pythonCongestionResponse `json:"congestion"`
	Trust         pythonTrustResponse      `json:"trust"`
	Targeting     pythonTargetingResponse  `json:"targeting"`
	ModelVersions map[string]string        `json:"model_versions"`
}

type pythonCongestionResponse struct {
	PredictedCongestionPenaltyMS float64 `json:"predicted_congestion_penalty_ms"`
	SaturationProbability        float64 `json:"saturation_probability"`
	Confidence                   float64 `json:"confidence"`
	ModelName                    string  `json:"model_name"`
}

type pythonTrustResponse struct {
	TrustScore float64 `json:"trust_score"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
	ModelName  string  `json:"model_name"`
}

type pythonTargetingResponse struct {
	TargetingRiskScore float64 `json:"targeting_risk_score"`
	Confidence         float64 `json:"confidence"`
	Reason             string  `json:"reason"`
	ModelName          string  `json:"model_name"`
}

// NewPythonModelClients creates model adapters backed by the Python prediction
// service. Set ML_SERVICE_URL to override the default http://127.0.0.1:8100.
func NewPythonModelClients(baseURL string) (*PythonCongestionClient, *PythonTrustClient, *PythonTargetingClient) {
	baseURL = PythonServiceURL(baseURL)
	store := &pythonPredictionStore{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 2 * time.Second},
		cache:   make(map[string]pythonCombinedResponse),
	}
	return &PythonCongestionClient{store: store}, &PythonTrustClient{store: store}, &PythonTargetingClient{store: store}
}

// PythonServiceURL returns the configured FastAPI service URL, applying the
// default used by the training service when no environment override is set.
func PythonServiceURL(baseURL string) string {
	if strings.TrimSpace(baseURL) == "" {
		return defaultPythonServiceURL
	}
	return strings.TrimRight(baseURL, "/")
}

// CheckPythonModelService verifies the FastAPI service is reachable before the
// desktop agent starts. This prevents silently running a different model path.
func CheckPythonModelService(baseURL string) error {
	baseURL = PythonServiceURL(baseURL)
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(baseURL + "/model-info")
	if err != nil {
		return fmt.Errorf("python ML service is required at %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("python ML service %s returned %s", baseURL, resp.Status)
	}

	var info struct {
		Congestion string `json:"congestion"`
		Trust      string `json:"trust"`
		Targeting  string `json:"targeting"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return fmt.Errorf("decode python ML service model-info: %w", err)
	}
	if info.Congestion == "" || info.Trust == "" || info.Targeting == "" {
		return fmt.Errorf("python ML service %s returned incomplete model-info", baseURL)
	}
	return nil
}

func (c *PythonCongestionClient) Predict(obs agent.LinkObservation) agent.CongestionPrediction {
	result, err := c.store.predictAll(obs)
	if err != nil {
		return agent.CongestionPrediction{Confidence: 0.0}
	}
	return agent.CongestionPrediction{
		PenaltyMS:             result.Congestion.PredictedCongestionPenaltyMS,
		SaturationProbability: clamp01(result.Congestion.SaturationProbability),
		Confidence:            clamp01(result.Congestion.Confidence),
	}
}

func (t *PythonTrustClient) Score(obs agent.LinkObservation, congestion agent.CongestionPrediction) agent.ScoreResult {
	result, err := t.store.predictAll(obs)
	if err != nil {
		return agent.ScoreResult{
			Score:      1.0,
			Confidence: 0.0,
			Reasons:    []string{fmt.Sprintf("Python ML service unavailable: %v", err)},
		}
	}
	return agent.ScoreResult{
		Score:      clamp01(result.Trust.TrustScore),
		Confidence: clamp01(result.Trust.Confidence),
		Reasons:    reasonList(result.Trust.Reason),
	}
}

func (t *PythonTargetingClient) Score(obs agent.LinkObservation, usage agent.UsageHistory) agent.ScoreResult {
	result, err := t.store.predictAll(obs)
	if err != nil {
		return agent.ScoreResult{
			Score:      0.0,
			Confidence: 0.0,
			Reasons:    []string{fmt.Sprintf("Python ML service unavailable: %v", err)},
		}
	}
	return agent.ScoreResult{
		Score:      clamp01(result.Targeting.TargetingRiskScore),
		Confidence: clamp01(result.Targeting.Confidence),
		Reasons:    reasonList(result.Targeting.Reason),
	}
}

func (s *pythonPredictionStore) predictAll(obs agent.LinkObservation) (pythonCombinedResponse, error) {
	key := cacheKey(obs)

	s.mu.Lock()
	if cached, ok := s.cache[key]; ok {
		s.mu.Unlock()
		return cached, nil
	}
	s.mu.Unlock()

	body, err := json.Marshal(obs)
	if err != nil {
		return pythonCombinedResponse{}, err
	}

	req, err := http.NewRequest(http.MethodPost, s.baseURL+"/predict/all", bytes.NewReader(body))
	if err != nil {
		return pythonCombinedResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return pythonCombinedResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return pythonCombinedResponse{}, fmt.Errorf("POST %s/predict/all returned %s", s.baseURL, resp.Status)
	}

	var result pythonCombinedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return pythonCombinedResponse{}, err
	}

	s.mu.Lock()
	if len(s.cache) > 512 {
		s.cache = make(map[string]pythonCombinedResponse)
	}
	s.cache[key] = result
	s.mu.Unlock()

	return result, nil
}

func cacheKey(obs agent.LinkObservation) string {
	self := "nil"
	if obs.SelfReportedLatencyMS != nil {
		self = fmt.Sprintf("%.6f", *obs.SelfReportedLatencyMS)
	}
	return fmt.Sprintf(
		"%d|%s|%.6f|%.6f|%.6f|%.6f|%s|%s",
		obs.Tick,
		obs.LinkID,
		obs.CapacityUnits,
		obs.CurrentLoad,
		obs.LoadRatio,
		obs.TrafficShare,
		self,
		obs.Status,
	)
}

func reasonList(reason string) []string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil
	}
	return []string{reason}
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
