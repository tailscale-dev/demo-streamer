#!/bin/bash

apt update
apt install -y golang

go install github.com/clstokes/go-http-streamer@latest
nohup ./go/bin/go-http-streamer &
