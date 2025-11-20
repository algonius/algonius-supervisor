package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/algonius/algonius-supervisor/pkg/a2a"
)

// DiscoveryClient provides methods for agent discovery
type DiscoveryClient struct {
	baseURL    string
	authToken  string
	httpClient HTTPClient
	timeout    time.Duration
}

// NewDiscoveryClient creates a new discovery client instance
func NewDiscoveryClient(baseURL, authToken string, httpClient HTTPClient) *DiscoveryClient {
	if httpClient == nil {
		httpClient = &DefaultHTTPClient{
			Timeout: 30 * time.Second,
		}
	}

	return &DiscoveryClient{
		baseURL:    baseURL,
		authToken:  authToken,
		httpClient: httpClient.(*DefaultHTTPClient),
		timeout:    30 * time.Second,
	}
}

// DiscoverAgents lists all available agents
func (d *DiscoveryClient) DiscoverAgents(ctx context.Context) ([]*a2a.A2AExtendedCard, error) {
	// Construct the request URL
	url := d.baseURL + "/discovery/agents"

	// Create the request
	req := &Request{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": "Bearer " + d.authToken,
			"Content-Type":  "application/json",
		},
	}

	// Make the request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make discovery request: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, a2a.NewA2AError(a2a.InvalidRequest, fmt.Sprintf("Discovery request failed with status: %d", resp.StatusCode))
	}

	// Parse the response
	var agentList struct {
		Agents []*a2a.A2AExtendedCard `json:"agents"`
		Total  int                    `json:"total"`
	}

	if err := json.Unmarshal(resp.Body, &agentList); err != nil {
		return nil, fmt.Errorf("failed to parse discovery response: %w", err)
	}

	return agentList.Agents, nil
}

// GetAgent retrieves detailed information about a specific agent
func (d *DiscoveryClient) GetAgent(ctx context.Context, agentID string) (*a2a.A2AExtendedCard, error) {
	// Construct the request URL
	url := d.baseURL + "/discovery/agents/" + agentID

	// Create the request
	req := &Request{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": "Bearer " + d.authToken,
			"Content-Type":  "application/json",
		},
	}

	// Make the request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make agent details request: %w", err)
	}

	// Check status code
	if resp.StatusCode == 404 {
		return nil, a2a.NewA2AError(a2a.AgentNotFoundError, fmt.Sprintf("Agent not found: %s", agentID))
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, a2a.NewA2AError(a2a.InvalidRequest, fmt.Sprintf("Agent details request failed with status: %d", resp.StatusCode))
	}

	// Parse the response
	var agentCard a2a.A2AExtendedCard
	if err := json.Unmarshal(resp.Body, &agentCard); err != nil {
		return nil, fmt.Errorf("failed to parse agent details response: %w", err)
	}

	return &agentCard, nil
}

// GetCapabilities retrieves the capabilities of the supervisor
func (d *DiscoveryClient) GetCapabilities(ctx context.Context) (map[string]interface{}, error) {
	// Construct the request URL
	url := d.baseURL + "/discovery/capabilities"

	// Create the request
	req := &Request{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": "Bearer " + d.authToken,
			"Content-Type":  "application/json",
		},
	}

	// Make the request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make capabilities request: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, a2a.NewA2AError(a2a.InvalidRequest, fmt.Sprintf("Capabilities request failed with status: %d", resp.StatusCode))
	}

	// Parse the response
	var capabilities map[string]interface{}
	if err := json.Unmarshal(resp.Body, &capabilities); err != nil {
		return nil, fmt.Errorf("failed to parse capabilities response: %w", err)
	}

	return capabilities, nil
}

// ResolveAgentCard resolves an agent card by querying the well-known endpoint
func (d *DiscoveryClient) ResolveAgentCard(ctx context.Context, agentID string) (*a2a.A2AExtendedCard, error) {
	// Construct the request URL for the well-known agent card endpoint
	url := d.baseURL + "/agents/" + agentID + "/v1/.well-known/agent-card.json"

	// Create the request without authentication as it's a public endpoint
	req := &Request{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// Make the request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve agent card: %w", err)
	}

	// Check status code
	if resp.StatusCode == 404 {
		return nil, a2a.NewA2AError(a2a.AgentNotFoundError, fmt.Sprintf("Agent card not found: %s", agentID))
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, a2a.NewA2AError(a2a.InvalidRequest, fmt.Sprintf("Agent card resolution failed with status: %d", resp.StatusCode))
	}

	// Parse the response
	var agentCard a2a.A2AExtendedCard
	if err := json.Unmarshal(resp.Body, &agentCard); err != nil {
		return nil, fmt.Errorf("failed to parse agent card: %w", err)
	}

	return &agentCard, nil
}

// ListAgentCards retrieves all available agent cards
func (d *DiscoveryClient) ListAgentCards(ctx context.Context) ([]*a2a.A2AExtendedCard, error) {
	// First, get the list of agents
	agents, err := d.DiscoverAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover agents: %w", err)
	}

	// Then resolve each agent's card
	cards := make([]*a2a.A2AExtendedCard, 0, len(agents))
	for _, agent := range agents {
		card, err := d.ResolveAgentCard(ctx, agent.ID)
		if err != nil {
			// Log the error but continue with other agents
			continue
		}
		cards = append(cards, card)
	}

	return cards, nil
}
