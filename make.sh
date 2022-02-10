#!/usr/bin/env bash
current_dir=`pwd`
go_grpc_dir=$current_dir"/grpc"
python_grpc_dir=$current_dir"/python/grpc"
echo "current_dir: $current_dir"
echo "go_grpc_dir: $go_grpc_dir"
echo "python_grpc_dir: $python_grpc_dir"
cd $current_dir
protoc --proto_path=./pb/ --go_out=plugins=grpc:./grpc common.proto
protoc --proto_path=./pb/ --go_out=plugins=grpc:./grpc bridge.proto
python -m grpc_tools.protoc --python_out=python/grpc --grpc_python_out=python/grpc --proto_path=./pb/ common.proto
python -m grpc_tools.protoc --python_out=python/grpc --grpc_python_out=python/grpc --proto_path=./pb/ bridge.proto
echo "Done"
