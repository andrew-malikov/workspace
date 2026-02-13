#!/usr/bin/env bash
#MISE description="Build the CLI"
#MISE tools={go="latest"}
go build -o binaries/ws main.go
