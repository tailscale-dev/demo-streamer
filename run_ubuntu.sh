#!/bin/bash

apt update
apt install -y golang

go install github.com/clstokes/streamer-demo@latest
nohup ./go/bin/streamer-demo &
