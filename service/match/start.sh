#!/usr/bin/env bash
set -e
COMMIT_HASH=$(git rev-parse --short HEAD || echo "GitNotFound")
COMMIT_DATE=$(git log --pretty=format:"%cd" -1 --date=format:"%Y-%m-%d %H:%M:%S")

echo "COMMIT_HASH:"${COMMIT_HASH}",COMMIT_DATE:"${COMMIT_DATE}


go build -o=match -mod vendor -ldflags "-X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=${COMMIT_DATE}\" " main.go
./match -config ./conf/main.yaml