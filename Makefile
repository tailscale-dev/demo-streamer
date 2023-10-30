build:
	docker build --tag clstokes/streamer-demo .

push:
	docker push clstokes/streamer-demo

run:
	docker run --publish 8080:8080 clstokes/streamer-demo
