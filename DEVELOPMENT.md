# NWDAF Development Notes

## Quick Start

```bash
# Install dependencies
go mod download

# Run locally
go run cmd/main.go -c config/nwdafcfg.yaml

# Build
make build

# Run with Docker
make docker-build
make docker-up
```

## Integration Points with free5GC

### 1. Network Registration
- NWDAF should register with NRF (Network Repository Function)
- Update `nrfUri` in config to point to your NRF instance

### 2. Data Collection
- Implement clients to collect data from:
  - AMF: UE registration, mobility events
  - SMF: Session management, QoS flows
  - UPF: Traffic statistics, throughput
  - PCF: Policy decisions

### 3. Analytics Consumers
- Other NFs (AMF, SMF, PCF) can subscribe to NWDAF analytics
- Implement notification callbacks to push analytics to consumers

## Next Steps

1. **Implement NRF Registration**
   - Add NRF client in `pkg/nrf/`
   - Register NWDAF services on startup
   - Implement heartbeat mechanism

2. **Add Data Collection Clients**
   - Implement HTTP clients for each NF type
   - Create data collection routines
   - Store collected data in context

3. **Enhance Analytics Engine**
   - Add ML models for predictions
   - Implement more sophisticated analytics
   - Add time-series data storage

4. **Add Persistence**
   - Database integration (MongoDB, InfluxDB)
   - Historical data storage
   - Analytics results caching

5. **Testing**
   - Unit tests for each package
   - Integration tests with mock NFs
   - Load testing for analytics engine

## Debugging

```bash
# Run with debug logging
./nwdaf -c config/nwdafcfg.yaml --loglevel debug

# Check specific logs
tail -f log/nwdaf.log | grep Analytics

# Test API endpoints
curl http://localhost:8000/health
```

## Performance Considerations

- Analytics engine runs periodically (configurable via `analyticsDelay`)
- Subscription processing is done in batches
- Consider adding caching for frequently requested analytics
- Use goroutines for parallel data collection from multiple NFs
