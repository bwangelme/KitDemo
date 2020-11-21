#!/usr/bin/env bash

# 安装 protobuf: https://github.com/protocolbuffers/protobuf#protocol-compiler-installation
# 安装 protoc-gen-go
#   为了和 Go-kit 兼容，使用 1.3.2 版本
#   go get github.com/golang/protobuf/protoc-gen-go@v1.3.2
#   go get github.com/golang/protobuf/proto@v1.3.2
protoc addsvc.proto --go_out=plugins=grpc:.