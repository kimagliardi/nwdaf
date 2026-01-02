package context

// SliceStatistics holds network slice statistics
type SliceStatistics struct {
	SNSSAI        string
	ActiveUEs     int
	Throughput    float64
	ResourceUsage float64
	Timestamp     int64
}
