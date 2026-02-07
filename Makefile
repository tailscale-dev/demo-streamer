BINARY_NAME=demo-streamer
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD)
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
DOCKER_REPO?=ghcr.io/tailscale-dev/demo-streamer
DOCKER_TAG?=$(VERSION)
DOCKER_PLATFORMS?=linux/amd64,linux/arm64
DOCKER_BUILD_PLATFORMS?=linux/$(shell docker info -f '{{.Architecture}}' 2>/dev/null || echo amd64)

LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

.PHONY: build
build:
	go build ${LDFLAGS} -o ${BINARY_NAME} .

.PHONY: dev
dev:
	go run ${LDFLAGS} main.go --dev

.PHONY: run
run:
	go run ${LDFLAGS} main.go

.PHONY: clean
clean:
	rm -f ${BINARY_NAME}

.PHONY: install
install:
	go install ${LDFLAGS} .

# Build with proper version for releases
.PHONY: release
release:
	@if [ -z "$(VERSION)" ]; then echo "VERSION is required"; exit 1; fi
	CGO_ENABLED=0 go build ${LDFLAGS} -o ${BINARY_NAME} .

# Docker targets
.PHONY: docker-build
docker-build:
	docker buildx build \
		--platform ${DOCKER_BUILD_PLATFORMS} \
		--build-arg VERSION=${VERSION} \
		--build-arg COMMIT=${COMMIT} \
		--build-arg DATE=${DATE} \
		-t ${DOCKER_REPO}:${DOCKER_TAG} \
		-t ${DOCKER_REPO}:latest \
		--load \
		.

.PHONY: docker-push
docker-push:
	docker buildx build \
		--platform ${DOCKER_PLATFORMS} \
		--build-arg VERSION=${VERSION} \
		--build-arg COMMIT=${COMMIT} \
		--build-arg DATE=${DATE} \
		-t ${DOCKER_REPO}:${DOCKER_TAG} \
		-t ${DOCKER_REPO}:latest \
		--push \
		.

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	@echo "Make sure to set TAILSCALE_AUTHKEY environment variable"
	docker run --rm -it \
		-e TAILSCALE_AUTHKEY=${TAILSCALE_AUTHKEY} \
		-e HOSTNAME=${HOSTNAME} \
		-e TLS=${TLS} \
		-e DEV=${DEV} \
		${DOCKER_REPO}:${DOCKER_TAG}

.PHONY: docker-run-dev
docker-run-dev:
	@echo "Running Docker container in development mode..."
	docker run --rm -it \
		-p 8080:8080 \
		-e DEV=true \
		-e TSNET=false \
		${DOCKER_REPO}:${DOCKER_TAG} \
		/app/demo-streamer --dev --port 8080

.PHONY: docker-clean
docker-clean:
	docker rmi ${DOCKER_REPO}:${DOCKER_TAG} ${DOCKER_REPO}:latest || true

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build           - Build the binary"
	@echo "  dev             - Run in development mode"
	@echo "  run             - Run the application"
	@echo "  clean           - Clean build artifacts"
	@echo "  install         - Install the binary"
	@echo "  release         - Build release binary (requires VERSION)"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-push     - Build and push multi-arch Docker image"
	@echo "  docker-run      - Run Docker container (tsnet mode)"
	@echo "  docker-run-dev  - Run Docker container (dev mode, port 8080)"
	@echo "  docker-clean    - Remove Docker images"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION         - Version tag (default: dev)"
	@echo "  DOCKER_REPO     - Docker repository (default: ghcr.io/jaxxstorm/tailscale-demo-streamer)"
	@echo "  DOCKER_TAG      - Docker tag (default: VERSION)"
	@echo "  DOCKER_PLATFORMS - Buildx platforms for docker-push (default: linux/amd64,linux/arm64)"
	@echo "  DOCKER_BUILD_PLATFORMS - Buildx platforms for docker-build (default: linux/<docker arch>)"
	@echo "  TAILSCALE_AUTHKEY - Tailscale auth key for tsnet mode"
	@echo "  HOSTNAME        - Custom hostname for tsnet"
	@echo "  TLS             - Enable/disable TLS (default: true)"
	@echo "  DEV             - Enable development mode"
