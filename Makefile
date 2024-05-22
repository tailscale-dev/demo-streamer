init:
	docker buildx create --name mybuilder --bootstrap --use

build:
	docker buildx build --tag tailscale-dev/demo-streamer --load .

push:
	docker buildx build --tag tailscale-dev/demo-streamer --platform linux/amd64,linux/arm64 --push .

run:
	docker run --rm --publish 80:80 tailscale-dev/demo-streamer
