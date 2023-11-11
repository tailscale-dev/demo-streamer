init:
	docker buildx create --name mybuilder --bootstrap --use

build:
	docker buildx build --tag clstokes/streamer-demo --load .

push:
	docker buildx build --tag clstokes/streamer-demo --platform linux/amd64,linux/arm64 --push .

run:
	docker run --rm --publish 8080:8080 clstokes/streamer-demo
