#!/usr/bin/env just --justfile

copy-css:
    cp node_modules/@picocss/pico/css/pico.classless.min.css internal/gensite

createdb-dev:
    docker build -t bitcoinrpc .
    docker rm -f bitcoinrpc || true
    docker run --name bitcoinrpc bitcoinrpc
    docker cp bitcoinrpc:/app/rpc.db .

update:
    go get -u
    go mod tidy -v