#!/bin/sh

build_server() {
    CMD_NAME=tora-server
    PACKAGE=./cmd/tora-server
    RELEASE_PATH=release/

    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $RELEASE_PATH/darwin/$CMD_NAME $PACKAGE
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $RELEASE_PATH/linux/$CMD_NAME $PACKAGE
    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -o $RELEASE_PATH/freebsd/$CMD_NAME $PACKAGE
}

build_cli() {
    CMD_NAME=tora-cli
    PACKAGE=./cmd/tora-cli
    RELEASE_PATH=release/

    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $RELEASE_PATH/darwin/$CMD_NAME $PACKAGE
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $RELEASE_PATH/linux/$CMD_NAME $PACKAGE
    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -o $RELEASE_PATH/freebsd/$CMD_NAME $PACKAGE
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $RELEASE_PATH/windows/$CMD_NAME $PACKAGE
}

rm -rf release/*
build_server
build_cli
tree -al ./release
