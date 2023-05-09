#!/bin/bash

apt update
apt install -y golang

go install github.com/clstokes/demo-streamer@latest
nohup ./go/bin/demo-streamer &
