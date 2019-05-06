#! /bin/bash

GOOS=windows GOARCH=amd64 go build -o build/geoipcli-go-windows-amd64
GOOS=darwin GOARCH=amd64 go build -o build/geoipcli-go-macos-amd64
GOOS=linux GOARCH=amd64 go build -o build/geoipcli-go-linux-amd64
