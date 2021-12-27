#!/usr/bin/env bash
current_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
go_grpc_dir=$current_dir"/../grpc"
python_grpc_dir=$current_dir"/../python/grpc"
echo "current_dir: $current_dir"
echo "go_grpc_dir: $go_grpc_dir"
echo "python_grpc_dir: $python_grpc_dir"
cd $current_dir
C:/Windows/System32/protoc --proto_path=./ --go_out=plugins=grpc:../grpc common.proto
C:/Windows/System32/protoc --proto_path=./ --go_out=plugins=grpc:../grpc go.proto
C:/Windows/System32/protoc --proto_path=./ --go_out=plugins=grpc:../grpc python.proto
G:/anaconda3/envs/DeepCast/python -m grpc_tools.protoc --python_out=$python_grpc_dir --grpc_python_out=$python_grpc_dir -I. common.proto
G:/anaconda3/envs/DeepCast/python -m grpc_tools.protoc --python_out=$python_grpc_dir --grpc_python_out=$python_grpc_dir -I. go.proto
G:/anaconda3/envs/DeepCast/python -m grpc_tools.protoc --python_out=$python_grpc_dir --grpc_python_out=$python_grpc_dir -I. python.proto
echo "Done"
