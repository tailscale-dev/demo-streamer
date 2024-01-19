#!/bin/bash

apt-get update
apt-get install -y golang

go install github.com/clstokes/demo-streamer@latest
nohup ./go/bin/demo-streamer &
