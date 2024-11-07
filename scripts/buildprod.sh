#!/bin/bash
cd cmd/api
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o out