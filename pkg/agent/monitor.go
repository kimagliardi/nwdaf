package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/free5gc/nwdaf/internal/logger"
)

// AutoSteeringMonitor logic
func (a *Agent) StartAutoSteerMonitor() {
	if !a.Config.AutoSteerEnabled {
		logger.AppLog.Infoln("⏸️  Auto-steering is disabled")
		return
	}

	logger.AppLog.Infof("🔍 LLM-driven auto-steering monitor started (interval: %ds)", a.Config.AutoSteerInterval)
	a.MonitorRunning = true

	go func() {
		ticker := time.NewTicker(time.Duration(a.Config.AutoSteerInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-a.ctx.Done():
				a.MonitorRunning = false
				return
			case <-ticker.C:
				a.monitorLoop()
			}
		}
	}()
}

func (a *Agent) monitorLoop() {
	// Get traffic rates
	rates, err := a.getUpfTrafficRates()
	if err != nil {
		logger.AppLog.Errorf("⚠️  Monitor error querying traffic: %v", err)
		return
	}

	// Update Prometheus metrics
	UpfTrafficRate.WithLabelValues("edge1").Set(rates["edge1"])
	UpfTrafficRate.WithLabelValues("edge2").Set(rates["edge2"])
	UpfTrafficRate.WithLabelValues("upfb").Set(rates["upfb"])

	// Check active policy
	activePolicy := a.getActivePolicy()
	var currentTarget float64
	if activePolicy == "edge1" {
		currentTarget = 1
	} else if activePolicy == "edge2" {
		currentTarget = 2
	}
	CurrentTarget.Set(currentTarget)

	logger.AppLog.Infof("📈 Traffic rates - edge1: %.1f KB/s, edge2: %.1f KB/s, upfb: %.1f KB/s (policy: %s)",
		rates["edge1"]/1000, rates["edge2"]/1000, rates["upfb"]/1000, activePolicy)

	// Ask LLM
	decision := a.askLlmForDecision(rates, activePolicy)
	if decision.ShouldSteer && decision.Target != "" {
		logger.AppLog.Infof("🤖 LLM Reasoning: %s", decision.Reason)
		a.executeSteering(decision.Target, decision.Reason)
	} else {
		logger.AppLog.Infof("🤖 LLM decision: No steering needed (%s) - Reasoning: %s", decision.Reason, decision.Reason)
	}
}

func (a *Agent) getUpfTrafficRates() (map[string]float64, error) {
	rates := map[string]float64{"edge1": 0, "edge2": 0, "upfb": 0}

	// Use a simpler query per-pod like in tools.go
	query := `rate(container_network_receive_bytes_total{namespace="free5gc",pod=~".*upf.*"}[1m])`

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/query?query=%s", a.Config.PrometheusUrl, url.QueryEscape(query)))
	if err != nil {
		return rates, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rates, fmt.Errorf("prometheus query failed: status %d", resp.StatusCode)
	}

	// Use flexible map-based unmarshaling
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return rates, fmt.Errorf("JSON decode error: %v", err)
	}

	status, _ := rawResponse["status"].(string)
	if status != "success" {
		return rates, fmt.Errorf("prometheus query failed: status=%s", status)
	}

	data, ok := rawResponse["data"].(map[string]interface{})
	if !ok {
		return rates, fmt.Errorf("invalid data field in response")
	}

	result, ok := data["result"].([]interface{})
	if !ok {
		return rates, fmt.Errorf("invalid result field in response")
	}

	for _, item := range result {
		r, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		metric, ok := r["metric"].(map[string]interface{})
		if !ok {
			continue
		}

		pod, _ := metric["pod"].(string)
		podLower := strings.ToLower(pod)

		// Get value
		value, ok := r["value"].([]interface{})
		if !ok || len(value) < 2 {
			continue
		}

		valStr, ok := value[1].(string)
		if !ok {
			continue
		}

		var val float64
		fmt.Sscanf(valStr, "%f", &val)

		// Aggregate by UPF type
		if strings.Contains(podLower, "upf1") || (strings.Contains(podLower, "anchor") && strings.Contains(podLower, "1")) {
			rates["edge1"] += val
		} else if strings.Contains(podLower, "upf2") || (strings.Contains(podLower, "anchor") && strings.Contains(podLower, "2")) {
			rates["edge2"] += val
		} else if strings.Contains(podLower, "upfb") {
			rates["upfb"] += val
		}
	}

	return rates, nil
}

func (a *Agent) getActivePolicy() string {
	baseUrl := fmt.Sprintf("%s/3gpp-traffic-influence/v1/%s/subscriptions", a.Config.NefUrl, a.Config.AfId)
	resp, err := http.Get(baseUrl)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var subs []struct {
			TrafficRoutes []struct {
				Dnai string `json:"dnai"`
			} `json:"trafficRoutes"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&subs); err == nil {
			for _, sub := range subs {
				for _, route := range sub.TrafficRoutes {
					dnai := strings.ToLower(route.Dnai)
					if dnai == "edge1" || dnai == "edge2" {
						return dnai
					}
				}
			}
		}
	}
	return ""
}

type SteeringDecision struct {
	ShouldSteer bool
	Target      string
	Reason      string
}

func (a *Agent) askLlmForDecision(rates map[string]float64, activePolicy string) SteeringDecision {
	// Check cooldown
	if time.Since(a.LastSteerTime).Seconds() < float64(a.Config.AutoSteerCooldown) {
		return SteeringDecision{false, "", "cooldown"}
	}

	edge1Kb := rates["edge1"] / 1000
	edge2Kb := rates["edge2"] / 1000
	upfbKb := rates["upfb"] / 1000
	thresholdKb := a.Config.AutoSteerThresholdBps / 1000

	prompt := fmt.Sprintf(`You are a 5G traffic steering expert. Based on these metrics, decide if traffic steering is needed.

METRICS:
- Edge1 traffic: %.2f KB/s
- Edge2 traffic: %.2f KB/s
- UPFB traffic: %.2f KB/s
- Active policy: %s
- Threshold: %.2f KB/s

RULES:
1. If no policy exists and UPFB > threshold: steer to the edge with LOWER traffic (if equal, choose edge1)
2. If policy exists and that edge > threshold: rebalance to the other edge (if it has 20%% less traffic)
3. Otherwise: no action needed

Respond with a JSON object:
{
  "reasoning": "Explain your logic here step-by-step...",
  "decision": "edge1" | "edge2" | "none"
}

IMPORTANT: If UPFB traffic exceeds threshold and no policy exists, you MUST choose edge1 or edge2, NOT none!

Your JSON response:`, edge1Kb, edge2Kb, upfbKb, activePolicy, thresholdKb)

	logger.AppLog.Infoln("🤖 Asking LLM for steering decision...")
	response, err := a.LLM.Query(prompt)
	if err != nil {
		logger.AppLog.Errorf("❌ LLM decision error: %v", err)
		return SteeringDecision{false, "", "error"}
	}

	// Try to find JSON in response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 {
		logger.AppLog.Errorf("❌ LLM returned invalid JSON: %s", response)
		return SteeringDecision{false, "", "invalid_json"}
	}

	jsonStr := response[start : end+1]
	var result struct {
		Reasoning string `json:"reasoning"`
		Decision  string `json:"decision"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		logger.AppLog.Errorf("❌ Failed to parse LLM JSON: %v", err)
		return SteeringDecision{false, "", "json_parse_error"}
	}

	decisionClean := strings.ToLower(strings.TrimSpace(result.Decision))
	if decisionClean == "edge1" {
		return SteeringDecision{true, "edge1", result.Reasoning}
	} else if decisionClean == "edge2" {
		return SteeringDecision{true, "edge2", result.Reasoning}
	}

	return SteeringDecision{false, "", result.Reasoning}
}

func (a *Agent) executeSteering(target, reason string) {
	oldTarget := "none"
	// We don't track old target easily here but we can approximate

	logger.AppLog.Infof("🚀 LLM-driven auto-steering: -> %s", target)

	// Execute using tool
	res, _ := a.SteerTraffic(target)
	if strings.Contains(res, "✅") {
		a.LastSteerTime = time.Now()
		AutoSteerTriggers.WithLabelValues(oldTarget, target, reason).Inc()
		logger.AppLog.Infof("✅ LLM auto-steer successful: now routing through %s", target)
	} else {
		logger.AppLog.Errorf("❌ LLM auto-steer failed: %s", res)
	}
}
