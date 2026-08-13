# --- Stage 1: Static Go binary build ---
FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o photoshow main.go  

# --- Stage 2: Final Runner Environment ---
FROM alpine:3.20
WORKDIR /app

# Install required networking layers
RUN apk update && apk add --no-cache \
    iptables \
    iproute2 \
    ca-certificates \
    curl \
    tailscale

# Pull binary from the builder layer
COPY --from=builder /app/photoshow /app/photoshow

# Copy and prepare our unified entrypoint orchestrator
COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE 8080 8081

ENTRYPOINT ["/app/entrypoint.sh"]
