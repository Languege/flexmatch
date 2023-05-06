#!/usr/bin/env bash

export PATH=bin/osx:$PATH

#编译客户端协议
protoc -I ./ ./result_code.proto  --go_out='./open'

#编译rpc协议
protoc \
-I ./ \
./match_rpc_service.proto \
--go_out=plugins=grpc:open