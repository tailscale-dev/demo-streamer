# syntax=docker/dockerfile:1

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

WORKDIR /app
COPY ui ./ui
COPY go.mod ./
COPY go.sum ./
COPY main.go ./

RUN go mod download
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /demo-streamer

# Final stage
FROM alpine:latest

# Install CA certificates
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /demo-streamer /app/

EXPOSE 8080
CMD ["/app/demo-streamer"]