#!/bin/bash

apt-get update
apt-get install -y golang

go install github.com/clstokes/streamer-demo@latest
nohup ./go/bin/streamer-demo &
