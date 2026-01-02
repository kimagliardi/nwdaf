package analytics

import (
	"testing"
	"context"
	"time"

	nwdafContext "github.com/free5gc/nwdaf/pkg/context"
)

func TestNewAnalyticsEngine(t *testing.T) {
	ctx := &nwdafContext.NWDAFContext{}
	engine := NewAnalyticsEngine(ctx)
	
	if engine == nil {
		t.Fatal("Expected engine to be created")
	}
	
	if engine.context != ctx {
		t.Error("Expected context to be set correctly")
	}
}

func TestGetAnalytics(t *testing.T) {
	ctx := nwdafContext.GetSelf()
	ctx.Init()
	
	engine := NewAnalyticsEngine(ctx)
	
	tests := []struct {
		name      string
		eventType string
		wantError bool
	}{
		{"NF Load", "NF_LOAD", false},
		{"Network Performance", "NETWORK_PERFORMANCE", false},
		{"Slice Load", "SLICE_LOAD", false},
		{"Unknown Type", "UNKNOWN", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.GetAnalytics(tt.eventType, nil)
			
			if (err != nil) != tt.wantError {
				t.Errorf("GetAnalytics() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if result == nil && tt.eventType != "UNKNOWN" {
				t.Error("Expected non-nil result for known event type")
			}
		})
	}
}

func TestAnalyticsEngineStart(t *testing.T) {
	ctx := nwdafContext.GetSelf()
	ctx.Init()
	
	engine := NewAnalyticsEngine(ctx)
	
	testCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	done := make(chan bool)
	go func() {
		engine.Start(testCtx)
		done <- true
	}()
	
	select {
	case <-done:
		// Engine stopped as expected
	case <-time.After(3 * time.Second):
		t.Error("Engine did not stop within timeout")
	}
}
