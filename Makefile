build:
	docker build --tag clstokes/streamer-demo .

push:
	docker push clstokes/streamer-demo

run:
	docker run -p 127.0.0.1:8080:8080/tcp clstokes/streamer-demo
