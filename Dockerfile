# syntax=docker/dockerfile:1

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /app
COPY ui ./ui
COPY go.mod ./
COPY go.sum ./
COPY main.go ./
COPY Makefile ./

RUN go mod download

# Build with version information
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH \
    go build -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o /demo-streamer

# Final stage
FROM alpine:latest

# Install CA certificates for TLS
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /demo-streamer /app/

# Environment variables for tsnet configuration
ENV TSNET=true
ENV TLS=true
ENV HOSTNAME=demo-streamer

# When using tsnet, we don't need to expose a specific port
# The application will register with Tailscale and be accessible via the tailnet

# Default command uses tsnet mode
CMD ["/app/demo-streamer", "--tsnet", "--tls"]