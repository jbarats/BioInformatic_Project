#!/usr/bin/env bash

export GOPATH=/usr/local/go
export PATH=/usr/local/go/bin

go build picoindex.go
go build picomap.go