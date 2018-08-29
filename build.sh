#!/bin/sh

export VERSION=unknown
export TAG=$(git describe $(git rev-list --tags --max-count=1))
if  [ -n "$1" ] ;then
  export VERSION=$1
else
  export VERSION=$TAG
fi

build_server() {
    CMD_NAME=tora-server
    PACKAGE=./cmd/tora-server
    RELEASE_PATH=release/

    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $RELEASE_PATH/darwin-${VERSION}-x64/$CMD_NAME $PACKAGE
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $RELEASE_PATH/linux-${VERSION}-x64/$CMD_NAME $PACKAGE
    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -o $RELEASE_PATH/freebsd-${VERSION}-x64/$CMD_NAME $PACKAGE
}

build_cli() {
    CMD_NAME=tora-cli
    PACKAGE=./cmd/tora-cli
    RELEASE_PATH=release/

    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $RELEASE_PATH/darwin-${VERSION}-x64/$CMD_NAME $PACKAGE
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $RELEASE_PATH/linux-${VERSION}-x64/$CMD_NAME $PACKAGE
    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -o $RELEASE_PATH/freebsd-${VERSION}-x64/$CMD_NAME $PACKAGE
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $RELEASE_PATH/windows-${VERSION}-x64/$CMD_NAME $PACKAGE
}

zip_pack() {
    cd release/$1
    zip -r ../$1.zip *
    cd ../..
}

echo "build version: ${VERSION}"
rm -rf release/*
build_server
build_cli
zip_pack darwin-${VERSION}-x64
zip_pack linux-${VERSION}-x64
zip_pack freebsd-${VERSION}-x64
zip_pack windows-${VERSION}-x64
tree -al ./release
