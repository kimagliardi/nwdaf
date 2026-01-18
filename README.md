# NWDAF - Network Data Analytics Function for free5GC

A Go implementation of the NWDAF (Network Data Analytics Function) for integration with [free5GC](https://github.com/free5gc/free5gc), an open-source 5G Core network implementation.

## Overview

NWDAF is a key network function in 5G networks that provides data analytics services to support intelligent network operations. This implementation provides:

- **Analytics Services**: Network performance, NF load, and slice analytics
- **Event Subscriptions**: Subscribe to analytics events and receive notifications
- **Data Collection**: Collect statistics from network functions (AMF, SMF, UPF, PCF)
- **Service-Based Interface (SBI)**: RESTful APIs following 3GPP TS 23.288 specifications

## Architecture

```
nwdaf/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── logger/                 # Logging functionality
│   └── sbi/                    # Service-Based Interface (HTTP APIs)
├── pkg/
│   ├── analytics/              # Analytics engine
│   ├── context/                # NWDAF context and data storage
│   ├── factory/                # Configuration factory
│   └── service/                # Main service logic
├── config/
│   └── nwdafcfg.yaml          # Configuration file
├── Dockerfile                  # Docker build configuration
├── docker-compose.yaml         # Docker Compose setup
└── Makefile                    # Build automation
```

## Features

### Supported Analytics Types

1. **NF_LOAD**: Network Function load analytics
   - Monitor and predict NF load levels
   - Detect overload conditions
   - Resource optimization insights

2. **NETWORK_PERFORMANCE**: Network-wide performance analytics
   - Latency, throughput, and packet loss metrics
   - Performance trend analysis
   - Quality of Service (QoS) monitoring

3. **SLICE_LOAD**: Network slice analytics
   - Slice resource utilization
   - Active UE tracking per slice
   - Slice performance metrics

### API Endpoints

#### Event Subscription Service (`/nnwdaf-eventssubscription/v1`)

- `POST /subscriptions` - Create analytics subscription
- `GET /subscriptions/:id` - Retrieve subscription details
- `PUT /subscriptions/:id` - Update subscription
- `DELETE /subscriptions/:id` - Delete subscription

#### Analytics Info Service (`/nnwdaf-analyticsinfo/v1`)

- `POST /analytics` - Request analytics data

#### Health Check

- `GET /health` - Service health status

## Installation

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (for containerized deployment)
- free5GC setup (optional, for integration)

### Local Build

```bash
# Clone the repository
git clone <your-repo-url>
cd nwdaf

# Download dependencies
make deps

# Build the binary
make build

# Run locally
make run
```

### Docker Build

```bash
# Build Docker image
make docker-build

# Start with Docker Compose
make docker-up

# View logs
make docker-logs

# Stop containers
make docker-down
```

### Kubernetes Deployment with Helm

This NWDAF can be deployed as part of the free5gc-helm charts. The Helm chart is located in `charts/free5gc-nwdaf/`.

#### Standalone Installation

```bash
# Create namespace
kubectl create ns free5gc

# Install NWDAF
helm -n free5gc install nwdaf ./charts/free5gc-nwdaf/

# Check the pod status
kubectl -n free5gc get pods -l "nf=nwdaf"

# Uninstall
helm -n free5gc uninstall nwdaf
```

#### Integration with free5gc-helm

To integrate NWDAF with the main free5gc-helm charts:

1. Copy the `charts/free5gc-nwdaf/` directory to `free5gc-helm/charts/free5gc/charts/`

2. Add NWDAF to the main chart's `Chart.yaml` dependencies:

```yaml
dependencies:
  # ... existing dependencies ...
  - name: free5gc-nwdaf
    version: "0.1.0"
    condition: deployNWDAF
```

3. Add to the main `values.yaml`:

```yaml
deployNWDAF: true

# Override NWDAF values if needed
free5gc-nwdaf:
  nwdaf:
    image:
      name: towards5gs/free5gc-nwdaf
```

4. Update dependencies and install:

```bash
cd free5gc-helm/charts/free5gc
helm dependency update
helm -n free5gc install free5gc .
```

## Configuration

Edit `config/nwdafcfg.yaml` to configure your NWDAF instance:

```yaml
configuration:
  nwdafName: NWDAF
  
  sbi:
    scheme: http
    registerIPv4: 127.0.0.10
    bindingIPv4: 0.0.0.0
    port: 8000

  nrfUri: http://127.0.0.10:8000
  
  plmnList:
    - mcc: "208"
      mnc: "93"
  
  serviceNameList:
    - nnwdaf-eventssubscription
    - nnwdaf-analyticsinfo

  analyticsDelay: 10  # Analytics computation interval (seconds)

  dataCollection:
    enabled: true
    collectionPeriod: 60  # Data collection interval (seconds)
    targetNFs:
      - AMF
      - SMF
      - UPF
      - PCF
```

## Integration with free5GC

### Add to free5gc-compose

1. Copy the NWDAF service definition to your `free5gc-compose/docker-compose.yaml`:

```yaml
nwdaf:
  container_name: nwdaf
  image: free5gc-nwdaf:latest
  build:
    context: ./nwdaf
    dockerfile: Dockerfile
  ports:
    - "8000:8000"
  volumes:
    - ./nwdaf/config:/root/config
  networks:
    privnet:
      ipv4_address: 10.100.200.20
```

2. Update network configuration to include NWDAF's IP address

3. Register NWDAF with NRF by setting the correct `nrfUri` in the config

## Usage Examples

### Create an Analytics Subscription

```bash
curl -X POST http://localhost:8000/nnwdaf-eventssubscription/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "eventType": "NF_LOAD",
    "consumerNfId": "amf-001",
    "notificationUri": "http://amf:8080/namf-callback/v1/nwdaf-notifications",
    "reportingPeriod": 60
  }'
```

### Request Analytics Data

```bash
curl -X POST http://localhost:8000/nnwdaf-analyticsinfo/v1/analytics \
  -H "Content-Type: application/json" \
  -d '{
    "eventType": "NETWORK_PERFORMANCE",
    "analyticsFilter": {
      "nfType": "SMF"
    }
  }'
```

### Delete a Subscription

```bash
curl -X DELETE http://localhost:8000/nnwdaf-eventssubscription/v1/subscriptions/{subscriptionId}
```

## Development

### Project Structure

- **cmd/**: Application entry point and CLI handling
- **internal/**: Private application code (logger, SBI handlers)
- **pkg/**: Public packages that can be imported
  - **analytics/**: Analytics engine and algorithms
  - **context/**: Context management and data storage
  - **factory/**: Configuration management
  - **service/**: Main service initialization and lifecycle

### Adding New Analytics Types

1. Define the analytics type in `pkg/analytics/engine.go`
2. Implement the analytics generation logic
3. Add corresponding handler in SBI API
4. Update configuration if needed

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -cover ./...
```

## Logging

Logs are categorized by component:

- **AppLog**: General application logs
- **InitLog**: Initialization logs
- **SbiLog**: SBI API request/response logs
- **AnalyticsLog**: Analytics engine logs
- **ContextLog**: Context management logs

Configure log level in `config/nwdafcfg.yaml`:

```yaml
logger:
  level: info  # trace|debug|info|warn|error|fatal|panic
  file: log/nwdaf.log  # Optional: log to file
```

## 3GPP Specifications

This implementation follows 3GPP specifications:

- **TS 23.288**: Architecture enhancements for 5G System (5GS) to support network data analytics services
- **TS 29.520**: Network Data Analytics Services (Nnwdaf)
- **TS 29.520**: Nnwdaf_EventsSubscription Service
- **TS 29.520**: Nnwdaf_AnalyticsInfo Service

## Roadmap

- [ ] Machine Learning integration for predictive analytics
- [ ] Support for additional analytics types (UE mobility, service experience)
- [ ] Integration with Prometheus for metrics export
- [ ] Enhanced data collection from NFs via Nnf services
- [ ] Support for multiple NWDAF instances (horizontal scaling)
- [ ] gRPC support for internal communication

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

[Specify your license here]

## Acknowledgments

- [free5GC](https://github.com/free5gc/free5gc) - Open source 5G Core Network
- 3GPP specifications for NWDAF architecture and interfaces

## Contact

[Your contact information or project maintainer details]
