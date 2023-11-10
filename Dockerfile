# syntax=docker/dockerfile:1

FROM golang:1.19-alpine

WORKDIR /app

COPY ui ./ui
COPY go.mod ./
COPY go.sum ./
COPY main.go ./

# RUN ls -al
# RUN ls -al ./ui

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /streamer-demo

EXPOSE 8080
CMD ["/streamer-demo"]
