# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /free5gc

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nwdaf ./cmd/main.go

# Runtime stage - using the same base as other free5gc NFs
FROM alpine:3.18

ENV GIN_MODE=release

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /free5gc

# Copy binary from builder
COPY --from=builder /free5gc/nwdaf ./

# Create config directory (config will be mounted via ConfigMap in k8s)
RUN mkdir -p config

# Copy default config for local development
COPY --from=builder /free5gc/config ./config

# Expose SBI port
EXPOSE 8000

# Set entrypoint
ENTRYPOINT ["./nwdaf"]
CMD ["-c", "config/nwdafcfg.yaml"]
