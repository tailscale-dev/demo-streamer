#!/bin/bash

apt-get update
apt-get install -y golang

go install github.com/tailscale-dev/demo-streamer@latest
nohup ./go/bin/demo-streamer &
