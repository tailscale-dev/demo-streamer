# streamer-demo

> :warning: Please don't use this for anything.
This is used as part of a demo for [Tailscale](https://tailscale.com/).

## Build and Run with Docker

```shell
docker build .
docker run streamer-demo
```

## Install and Run on Ubuntu

```shell
apt update
apt install -y golang

go install github.com/clstokes/streamer-demo@latest
nohup ./go/bin/streamer-demo &
```

or

```shell
curl -fsSL https://raw.githubusercontent.com/clstokes/streamer-demo/main/run_ubuntu.sh | sh
```

## To enable Tailscale Funnel

```shell
tailscale serve https / http://127.0.0.1:8080
tailscale funnel 443 on
```
