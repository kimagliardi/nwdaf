package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/free5gc/nwdaf/internal/logger"
)

// Tool 1: Get UPF Network Metrics from Prometheus
func (a *Agent) GetUPFNetworkMetrics() (string, error) {
	// Query for network bytes received/transmited by UPF pods
	rxQuery := `container_network_receive_bytes_total{namespace="free5gc",pod=~".*upf.*"}`
	txQuery := `container_network_transmit_bytes_total{namespace="free5gc",pod=~".*upf.*"}`
	rxRateQuery := `rate(container_network_receive_bytes_total{namespace="free5gc",pod=~".*upf.*"}[1m])`
	txRateQuery := `rate(container_network_transmit_bytes_total{namespace="free5gc",pod=~".*upf.*"}[1m])`

	results := make(map[string]map[string]interface{})

	// Helper to perform query and update results
	queryPrometheus := func(query string, metricName string) error {
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/query?query=%s", a.Config.PrometheusUrl, query))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("prometheus query failed: %d", resp.StatusCode)
		}

		var data struct {
			Data struct {
				Result []struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				} `json:"result"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return err
		}

		for _, res := range data.Data.Result {
			pod := res.Metric["pod"]
			iface := res.Metric["interface"]
			if pod == "" || iface == "" {
				continue
			}

			key := fmt.Sprintf("%s|%s", pod, iface)
			if _, ok := results[key]; !ok {
				results[key] = map[string]interface{}{
					"pod":       pod,
					"interface": iface,
				}
			}

			// Value is [timestamp, string_value]
			if len(res.Value) >= 2 {
				valStr, ok := res.Value[1].(string)
				if ok {
					var val float64
					fmt.Sscanf(valStr, "%f", &val)
					results[key][metricName] = val
				}
			}
		}
		return nil
	}

	if err := queryPrometheus(rxQuery, "rx_bytes"); err != nil {
		return "", fmt.Errorf("failed to query RX bytes: %v", err)
	}
	if err := queryPrometheus(txQuery, "tx_bytes"); err != nil {
		return "", fmt.Errorf("failed to query TX bytes: %v", err)
	}
	if err := queryPrometheus(rxRateQuery, "rx_rate"); err != nil {
		return "", fmt.Errorf("failed to query RX rate: %v", err)
	}
	if err := queryPrometheus(txRateQuery, "tx_rate"); err != nil {
		return "", fmt.Errorf("failed to query TX rate: %v", err)
	}

	if len(results) == 0 {
		return "No UPF network metrics found in Prometheus.", nil
	}

	// Format table
	var sb strings.Builder
	sb.WriteString(strings.Repeat("=", 100) + "\n")
	sb.WriteString("UPF Network Usage (from Prometheus)\n")
	sb.WriteString(strings.Repeat("=", 100) + "\n")
	sb.WriteString(fmt.Sprintf("%-45s %-12s %-12s %-12s %-12s %-12s\n", "POD", "INTERFACE", "RX TOTAL", "TX TOTAL", "RX RATE", "TX RATE"))
	sb.WriteString(strings.Repeat("-", 100) + "\n")

	// Sort keys
	var keys []string
	for k := range results {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		r := results[k]
		pod := r["pod"].(string)
		if len(pod) > 44 {
			pod = pod[:44]
		}
		iface := r["interface"].(string)
		if len(iface) > 11 {
			iface = iface[:11]
		}

		rxBytes := formatBytes(getFloat(r, "rx_bytes"))
		txBytes := formatBytes(getFloat(r, "tx_bytes"))
		rxRate := formatRate(getFloat(r, "rx_rate"))
		txRate := formatRate(getFloat(r, "tx_rate"))

		sb.WriteString(fmt.Sprintf("%-45s %-12s %-12s %-12s %-12s %-12s\n", pod, iface, rxBytes, txBytes, rxRate, txRate))
	}
	sb.WriteString(strings.Repeat("=", 100) + "\n")

	return sb.String(), nil
}

// Tool 2: Steer Traffic via NEF API
func (a *Agent) SteerTraffic(target string) (string, error) {
	target = strings.ToLower(strings.TrimSpace(target))
	if target != "edge1" && target != "edge2" {
		return fmt.Sprintf("❌ Invalid target: '%s'. Must be 'edge1' or 'edge2'", target), nil
	}

	baseUrl := fmt.Sprintf("%s/3gpp-traffic-influence/v1/%s/subscriptions", a.Config.NefUrl, a.Config.AfId)

	// Step 1: Delete existing subscriptions
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err == nil {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			var subs []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&subs); err == nil {
				for _, sub := range subs {
					self, ok := sub["self"].(string)
					if ok {
						parts := strings.Split(self, "/")
						subId := parts[len(parts)-1]
						delReq, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", baseUrl, subId), nil)
						client.Do(delReq)
					}
				}
				time.Sleep(1 * time.Second)
			}
			resp.Body.Close()
		}
	}

	// Step 2: Create new subscription
	payload := map[string]interface{}{
		"afServiceId": "steering",
		"afAppId":     "traffic-steering-agent",
		"dnn":         a.Config.Dnn,
		"snssai": map[string]interface{}{
			"sst": a.Config.Sst,
			"sd":  a.Config.Sd,
		},
		"anyUeInd": true,
		"trafficFilters": []map[string]interface{}{
			{
				"flowId":           1,
				"flowDescriptions": []string{"permit out ip from any to any"},
			},
		},
		"trafficRoutes": []map[string]interface{}{
			{
				"dnai": target,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err = http.NewRequest("POST", baseUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("❌ Cannot connect to NEF at %s", a.Config.NefUrl), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		var data map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&data)

		subId := ""
		if self, ok := data["self"].(string); ok {
			parts := strings.Split(self, "/")
			subId = parts[len(parts)-1]
		}

		upfName := "AnchorUPF1"
		pool := "10.1.0.0/17"
		if target == "edge2" {
			upfName = "AnchorUPF2"
			pool = "10.1.128.0/17"
		}

		logger.AppLog.Infof("Traffic steering created: target=%s subId=%s", target, subId)

		return fmt.Sprintf(`✅ Traffic steering subscription created!

Target DNAI: %s
Target UPF: %s
Expected IP Pool: %s
Subscription ID: %s

Traffic will be routed through %s.`, target, upfName, pool, subId, upfName), nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Sprintf("❌ Failed to create subscription: HTTP %d - %s", resp.StatusCode, string(body)), nil
}

// Helpers
func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		return v.(float64)
	}
	return 0
}

func formatBytes(b float64) string {
	if b >= 1e9 {
		return fmt.Sprintf("%.2f GB", b/1e9)
	} else if b >= 1e6 {
		return fmt.Sprintf("%.2f MB", b/1e6)
	} else if b >= 1e3 {
		return fmt.Sprintf("%.2f KB", b/1e3)
	}
	return fmt.Sprintf("%.0f B", b)
}

func formatRate(r float64) string {
	if r >= 1e6 {
		return fmt.Sprintf("%.2f Mbps", r*8/1e6)
	} else if r >= 1e3 {
		return fmt.Sprintf("%.2f Kbps", r*8/1e3)
	}
	return fmt.Sprintf("%.0f bps", r*8)
}
