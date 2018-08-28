#!/bin/sh

build_server() {
    PACKAGE=./cmd/tora-server
    RELEASE_PATH=release/tora-server

    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $RELEASE_PATH-darwin $PACKAGE
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $RELEASE_PATH-linux $PACKAGE
    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -o $RELEASE_PATH-freebsd $PACKAGE
}

build_cli() {
    PACKAGE=./cmd/tora-cli
    RELEASE_PATH=release/tora-cli

    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $RELEASE_PATH-darwin $PACKAGE
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $RELEASE_PATH-linux $PACKAGE
    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -o $RELEASE_PATH-freebsd $PACKAGE
}

rm -rf release/*
build_server
build_cli
ls -al ./release
