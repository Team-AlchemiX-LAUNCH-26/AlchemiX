package orchestrator

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/launch26/relic-ring-protocol/internal/domain"
	"github.com/launch26/relic-ring-protocol/internal/transport"
	"github.com/launch26/relic-ring-protocol/pkg/protocol"
)

type RemoteClient struct {
	BaseURL string
	client  *transport.Client
}

func NewRemoteClient(baseURL string) *RemoteClient {
	return &RemoteClient{BaseURL: strings.TrimRight(baseURL, "/"), client: transport.NewClient(20 * time.Second)}
}
func (c *RemoteClient) Universe(ctx context.Context) (protocol.UniverseResponse, error) {
	var out protocol.UniverseResponse
	err := c.client.GetJSON(ctx, c.BaseURL+"/api/universe", &out)
	return out, err
}
func (c *RemoteClient) Transmit(ctx context.Context, req protocol.TransmissionRequest) (domain.TransmissionResult, error) {
	var out domain.TransmissionResult
	err := c.client.PostJSON(ctx, c.BaseURL+"/api/transmissions", req, &out)
	return out, err
}
func (c *RemoteClient) Action(ctx context.Context, path string, body any) error {
	return c.client.PostJSON(ctx, c.BaseURL+path, body, nil)
}
func (c *RemoteClient) StreamEvents(ctx context.Context, onEvent func(protocol.Event)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/events", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("event stream returned HTTP %d", resp.StatusCode)
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var event protocol.Event
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err == nil {
			onEvent(event)
		}
	}
	return scanner.Err()
}
