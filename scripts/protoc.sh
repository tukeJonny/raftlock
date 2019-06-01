#!/usr/bin/env bash

pushd ./pb
protoc -I. --go_out=plugins=grpc:. raftlock.proto
popd