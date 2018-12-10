#! /bin/bash

GOOS=darwin GOARCH=amd64 go build -o geoipcli-go-macos-amd64
GOOS=linux GOARCH=amd64 go build -o geoipcli-go-linux-amd64
