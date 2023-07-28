build:
	docker build --tag clstokes/demo-streamer .

push:
	docker push clstokes/demo-streamer

run:
	docker run -p 127.0.0.1:8080:8080/tcp clstokes/demo-streamer
