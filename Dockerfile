# syntax=docker/dockerfile:1

FROM golang:1.21-alpine

WORKDIR /app

COPY ui ./ui
COPY go.mod ./
COPY go.sum ./
COPY main.go ./

# RUN ls -al
# RUN ls -al ./ui

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /demo-streamer

EXPOSE 80
CMD ["/demo-streamer"]
