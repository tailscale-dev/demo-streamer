init:
	docker buildx create --name mybuilder --bootstrap --use

build:
	docker buildx build --tag clstokes/demo-streamer --load .

push:
	docker buildx build --tag clstokes/demo-streamer --platform linux/amd64,linux/arm64 --push .

run:
	docker run --rm --publish 8080:8080 clstokes/demo-streamer
