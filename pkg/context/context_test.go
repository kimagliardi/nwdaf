package context

import (
	"testing"
)

func TestGetSelf(t *testing.T) {
	ctx := GetSelf()
	if ctx == nil {
		t.Fatal("Expected context to be initialized")
	}
}

func TestAddSubscription(t *testing.T) {
	ctx := GetSelf()
	
	sub := &AnalyticsSubscription{
		SubscriptionId:  "test-sub-1",
		EventType:       "NF_LOAD",
		ConsumerNfId:    "amf-001",
		NotificationUri: "http://amf:8080/callback",
	}
	
	ctx.AddSubscription(sub)
	
	retrieved, ok := ctx.GetSubscription("test-sub-1")
	if !ok {
		t.Error("Expected subscription to be found")
	}
	
	if retrieved.SubscriptionId != sub.SubscriptionId {
		t.Errorf("Expected subscription ID %s, got %s", sub.SubscriptionId, retrieved.SubscriptionId)
	}
}

func TestRemoveSubscription(t *testing.T) {
	ctx := GetSelf()
	
	sub := &AnalyticsSubscription{
		SubscriptionId:  "test-sub-2",
		EventType:       "NF_LOAD",
		ConsumerNfId:    "amf-001",
		NotificationUri: "http://amf:8080/callback",
	}
	
	ctx.AddSubscription(sub)
	ctx.RemoveSubscription("test-sub-2")
	
	_, ok := ctx.GetSubscription("test-sub-2")
	if ok {
		t.Error("Expected subscription to be removed")
	}
}

func TestUpdateNFStatistics(t *testing.T) {
	ctx := GetSelf()
	
	stats := &NFStatistics{
		NFInstanceId: "nf-001",
		NFType:       "AMF",
		Load:         0.75,
		Timestamp:    1234567890,
	}
	
	ctx.UpdateNFStatistics("nf-001", stats)
	
	retrieved, ok := ctx.GetNFStatistics("nf-001")
	if !ok {
		t.Error("Expected NF statistics to be found")
	}
	
	if retrieved.Load != stats.Load {
		t.Errorf("Expected load %.2f, got %.2f", stats.Load, retrieved.Load)
	}
}

func TestGetAllNFStatistics(t *testing.T) {
	ctx := GetSelf()
	
	stats1 := &NFStatistics{
		NFInstanceId: "nf-001",
		NFType:       "AMF",
		Load:         0.75,
	}
	
	stats2 := &NFStatistics{
		NFInstanceId: "nf-002",
		NFType:       "SMF",
		Load:         0.60,
	}
	
	ctx.UpdateNFStatistics("nf-001", stats1)
	ctx.UpdateNFStatistics("nf-002", stats2)
	
	allStats := ctx.GetAllNFStatistics()
	
	if len(allStats) < 2 {
		t.Errorf("Expected at least 2 NF statistics, got %d", len(allStats))
	}
}
