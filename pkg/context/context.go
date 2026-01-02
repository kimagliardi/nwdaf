package context

import (
	"sync"

	"github.com/free5gc/nwdaf/pkg/factory"
)

var nwdafContext *NWDAFContext
var nwdafContextOnce sync.Once

type NWDAFContext struct {
	NfId          string
	Name          string
	UriScheme     string
	BindingIPv4   string
	RegisterIPv4  string
	SBIPort       int
	NrfUri        string
	
	// Analytics subscriptions
	Subscriptions map[string]*AnalyticsSubscription
	SubMutex      sync.RWMutex
	
	// Data storage
	DataStore     *DataStore
	DataMutex     sync.RWMutex
}

type AnalyticsSubscription struct {
	SubscriptionId    string
	EventType         string
	ConsumerNfId      string
	NotificationUri   string
	AnalyticsFilter   map[string]interface{}
	ReportingPeriod   int
}

type DataStore struct {
	// Network function statistics
	NFStats       map[string]*NFStatistics
	
	// UE statistics
	UEStats       map[string]*UEStatistics
	
	// Slice statistics
	SliceStats    map[string]*SliceStatistics
}

type NFStatistics struct {
	NFInstanceId  string
	NFType        string
	Load          float64
	Timestamp     int64
	Metrics       map[string]float64
}

type UEStatistics struct {
	SUPI          string
	Location      string
	Throughput    float64
	Latency       float64
	PacketLoss    float64
	Timestamp     int64
}

type SliceStats struct {
	SNSSAI        string
	ActiveUEs     int
	Throughput    float64
	ResourceUsage float64
	Timestamp     int64
}

func init() {
	nwdafContextOnce.Do(func() {
		nwdafContext = &NWDAFContext{
			Subscriptions: make(map[string]*AnalyticsSubscription),
			DataStore:     NewDataStore(),
		}
	})
}

func GetSelf() *NWDAFContext {
	return nwdafContext
}

func (c *NWDAFContext) Init() {
	config := factory.NwdafConfig.Configuration

	c.Name = config.NwdafName
	c.UriScheme = config.Sbi.Scheme
	c.RegisterIPv4 = config.Sbi.RegisterIPv4
	c.BindingIPv4 = config.Sbi.BindingIPv4
	c.SBIPort = config.Sbi.Port
	c.NrfUri = config.NrfUri
}

func NewDataStore() *DataStore {
	return &DataStore{
		NFStats:    make(map[string]*NFStatistics),
		UEStats:    make(map[string]*UEStatistics),
		SliceStats: make(map[string]*SliceStatistics),
	}
}

func (c *NWDAFContext) AddSubscription(sub *AnalyticsSubscription) {
	c.SubMutex.Lock()
	defer c.SubMutex.Unlock()
	c.Subscriptions[sub.SubscriptionId] = sub
}

func (c *NWDAFContext) RemoveSubscription(subId string) {
	c.SubMutex.Lock()
	defer c.SubMutex.Unlock()
	delete(c.Subscriptions, subId)
}

func (c *NWDAFContext) GetSubscription(subId string) (*AnalyticsSubscription, bool) {
	c.SubMutex.RLock()
	defer c.SubMutex.RUnlock()
	sub, ok := c.Subscriptions[subId]
	return sub, ok
}

func (c *NWDAFContext) UpdateNFStatistics(nfId string, stats *NFStatistics) {
	c.DataMutex.Lock()
	defer c.DataMutex.Unlock()
	c.DataStore.NFStats[nfId] = stats
}

func (c *NWDAFContext) UpdateUEStatistics(supi string, stats *UEStatistics) {
	c.DataMutex.Lock()
	defer c.DataMutex.Unlock()
	c.DataStore.UEStats[supi] = stats
}

func (c *NWDAFContext) GetNFStatistics(nfId string) (*NFStatistics, bool) {
	c.DataMutex.RLock()
	defer c.DataMutex.RUnlock()
	stats, ok := c.DataStore.NFStats[nfId]
	return stats, ok
}

func (c *NWDAFContext) GetAllNFStatistics() map[string]*NFStatistics {
	c.DataMutex.RLock()
	defer c.DataMutex.RUnlock()
	
	result := make(map[string]*NFStatistics)
	for k, v := range c.DataStore.NFStats {
		result[k] = v
	}
	return result
}
