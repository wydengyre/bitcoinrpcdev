#!/usr/bin/env just --justfile

ci-test: install-node-deps copy-css test

copy-css:
    cp node_modules/@picocss/pico/css/pico.min.css internal/gensite

# run createdb within docker and copy out the produced DB. Very handy on MacOS.
createdb-dev:
    docker build -t bitcoinrpc .
    docker rm -f bitcoinrpc || true
    docker run --name bitcoinrpc bitcoinrpc
    docker cp bitcoinrpc:/app/rpc.db .

install-node-deps:
    npm install --frozen-lockfile

lint:
    golangci-lint run

test:
    go test -mod=readonly -v ./...

update:
    go get -u
    go mod tidy -v