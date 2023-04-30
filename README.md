# go-http-streamer

> :warning: Please don't use this for anything.
This is used as part of a demo for [Tailscale](https://tailscale.com/).

## Install and Run on Ubuntu

```shell
apt update
apt install -y golang

go install github.com/clstokes/go-http-streamer@latest
nohup ./go/bin/go-http-streamer &
```

or

```shell
curl -fsSL https://raw.githubusercontent.com/clstokes/go-http-streamer/main/run_ubuntu.sh | sh
```

## To enable Tailscale Funnel

```shell
tailscale serve https / http://127.0.0.1:8080
tailscale funnel 443 on
```
