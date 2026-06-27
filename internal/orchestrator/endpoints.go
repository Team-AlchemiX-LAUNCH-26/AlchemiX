package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func ResolveNodeEndpoints(cfg domain.UniverseConfig) (map[string]string, error) {
	if raw := strings.TrimSpace(os.Getenv("NODE_ENDPOINTS")); raw != "" {
		var endpoints map[string]string
		if err := json.Unmarshal([]byte(raw), &endpoints); err != nil {
			return nil, fmt.Errorf("decode NODE_ENDPOINTS: %w", err)
		}
		return endpoints, nil
	}
	endpoints := make(map[string]string, len(cfg.Nodes))
	for i, p := range cfg.Nodes {
		endpoints[p.ID] = fmt.Sprintf("http://localhost:%d", 8101+i)
	}
	return endpoints, nil
}
