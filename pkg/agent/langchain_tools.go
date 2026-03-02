package agent

import (
	"context"
	"encoding/json"
	"strings"
)

// MetricsTool implements the LangChainGo Tool interface for getting UPF metrics
type MetricsTool struct {
	agent *Agent
}

func NewMetricsTool(agent *Agent) *MetricsTool {
	return &MetricsTool{agent: agent}
}

func (t *MetricsTool) Name() string {
	return "get_upf_network_metrics"
}

func (t *MetricsTool) Description() string {
	return `Get current UPF network metrics from Prometheus. 
Returns traffic statistics including RX/TX bytes and rates for all UPF pods.
Use this tool when you need to check network traffic, analyze UPF performance, 
or make decisions about traffic steering.
No input required - just call the tool.`
}

func (t *MetricsTool) Call(ctx context.Context, input string) (string, error) {
	result, err := t.agent.GetUPFNetworkMetrics()
	if err != nil {
		return "Error getting metrics: " + err.Error(), nil
	}
	return result, nil
}

// SteerTrafficTool implements the LangChainGo Tool interface for steering traffic
type SteerTrafficTool struct {
	agent *Agent
}

func NewSteerTrafficTool(agent *Agent) *SteerTrafficTool {
	return &SteerTrafficTool{agent: agent}
}

func (t *SteerTrafficTool) Name() string {
	return "steer_traffic"
}

func (t *SteerTrafficTool) Description() string {
	return `Steer 5G traffic to a specific edge UPF via the NEF API.
Input should be the target edge: either "edge1" or "edge2".
- edge1: Routes traffic through AnchorUPF1 (IP pool 10.1.0.0/17)
- edge2: Routes traffic through AnchorUPF2 (IP pool 10.1.128.0/17)
Use this tool when you need to redirect traffic, balance load between edges,
or respond to high traffic conditions.`
}

func (t *SteerTrafficTool) Call(ctx context.Context, input string) (string, error) {
	// Parse the input - it might be JSON or plain string
	target := strings.TrimSpace(input)

	// Try to parse as JSON in case the LLM wraps it
	var jsonInput struct {
		Target string `json:"target"`
	}
	if err := json.Unmarshal([]byte(input), &jsonInput); err == nil && jsonInput.Target != "" {
		target = jsonInput.Target
	}

	// Clean up the target
	target = strings.ToLower(strings.Trim(target, `"' `))

	result, err := t.agent.SteerTraffic(target)
	if err != nil {
		return "Error steering traffic: " + err.Error(), nil
	}
	return result, nil
}
