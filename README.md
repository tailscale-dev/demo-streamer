# demo-streamer

> :warning: Please don't use this for anything.
This is used as part of a demo for [Tailscale](https://tailscale.com/).

## Features

- **tsnet Integration**: Run as a Tailscale node directly without requiring a separate Tailscale client
- **Automatic TLS**: Get free HTTPS certificates automatically via Tailscale
- **User Identity**: Shows logged-in Tailscale user information
- **Prometheus Metrics**: Built-in metrics collection
- **Flexible Deployment**: Can run in traditional mode or as a tsnet node

## Quick Start with Docker and tsnet

The easiest way to run this demo is using Docker with tsnet mode:

1. **Get a Tailscale auth key** from https://login.tailscale.com/admin/settings/keys

2. **Copy the environment template**:
   ```shell
   cp .env.example .env
   # Edit .env and add your TAILSCALE_AUTHKEY
   ```

3. **Run with Docker Compose**:
   ```shell
   docker-compose up -d
   ```

The application will:
- Register itself as a node in your Tailscale network
- Generate and manage TLS certificates automatically
- Be accessible at `https://demo-streamer.<your-tailnet>.ts.net` (or your custom hostname)

## Configuration Options

### Environment Variables / Command Line Flags

| Environment Variable | Flag | Default | Description |
|---------------------|------|---------|-------------|
| `PORT` | `--port` | `8080` | Port to listen on (traditional mode only) |
| `DEV` | `--dev` | `false` | Enable development mode |
| `TSNET` | `--tsnet` | `false` | Enable tsnet mode for Tailscale integration |
| `HOSTNAME` | `--hostname` | `tailscale-demo-streamer` | Hostname for tsnet registration |
| `TAILSCALE_AUTHKEY` | `--auth-key` | | Tailscale auth key for tsnet |
| `TLS` | `--tls` | `true` | Enable TLS certificate generation (tsnet mode) |

### Example Commands

**Traditional mode** (requires Tailscale client installed):
```shell
./demo-streamer --port 8080
```

**tsnet mode** (registers as Tailscale node directly):
```shell
./demo-streamer --tsnet --auth-key=tskey-auth-your-key --hostname=my-demo
```

**Development mode** (serves assets from filesystem):
```shell
./demo-streamer --dev --tsnet --auth-key=tskey-auth-your-key
```

## Build and Run Options

### Local Development

```shell
# Install dependencies
go mod download

# Run in development mode
make dev

# Build binary
make build

# Run with custom version info
make VERSION=1.0.0 build
```

### Docker Build

**Traditional mode**:
```shell
docker build --tag demo-streamer .
docker run --publish 8080:8080 demo-streamer /app/demo-streamer --port 8080
```

**tsnet mode**:
```shell
docker build --tag demo-streamer .
docker run -e TAILSCALE_AUTHKEY=your-auth-key demo-streamer
```

### Docker Compose

```shell
# Copy and edit environment variables
cp .env.example .env

# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

## Install and Run on Ubuntu

```shell
apt-get update
apt-get install -y golang

go install github.com/tailscale-dev/demo-streamer@latest
nohup ./go/bin/demo-streamer &
```

or

```shell
curl -fsSL https://raw.githubusercontent.com/tailscale-dev/demo-streamer/main/run_ubuntu.sh | sh
```

## Traditional Tailscale Setup (non-tsnet)

If you prefer to use the traditional Tailscale client instead of tsnet:

1. Install and configure Tailscale on your system
2. Run the application in traditional mode:
   ```shell
   ./demo-streamer --port 8080
   ```
3. Enable Tailscale Serve/Funnel:
   ```shell
   tailscale serve https / http://127.0.0.1:8080
   tailscale funnel 443 on
   ```

## Endpoints

- `/` - Main application with user identity display
- `/api/uuid` - Generate a random UUID
- `/metrics` - Prometheus metrics
- `/ui/*` - Static assets

## Version Information

The application includes build information that helps with debugging the "ERR-BuildInfo" issue in Tailscale console:

```shell
./demo-streamer --version
```
