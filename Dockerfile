# syntax=docker/dockerfile:1

FROM golang:1.19-alpine

WORKDIR /app

COPY src/ ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /demo-streamer

EXPOSE 8080
CMD ["/demo-streamer"]
