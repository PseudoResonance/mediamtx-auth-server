#!/usr/bin/env sh

mkdir $PWD/dist

go build -o $PWD/dist/server -v ./internal/server

echo "Starting..."

$PWD/dist/server "$@"
