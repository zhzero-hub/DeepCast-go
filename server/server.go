package server

import (
	grpc2 "DeepCast/grpc"
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	// Address 监听地址
	Address string = "127.0.0.1:5001"
	// Network 网络通信协议
	Network string = "tcp"
)

type GoServer struct {
	grpc2.ServerApiServer
}

func (*GoServer) GoSayHello(ctx context.Context, request *grpc2.GoSayHelloRequest) (*grpc2.GoSayHelloResponse, error) {
	return &grpc2.GoSayHelloResponse{
		Msg: "hello " + request.Msg,
	}, nil
}

func (*GoServer) NotifyAction(ctx context.Context, request *grpc2.NotifyActionRequest) (*grpc2.NotifyActionResponse, error) {
	log.Printf("receive notify action request: %v", request)
	return &grpc2.NotifyActionResponse{}, nil
}

func StartGoServer() {
	// 监听本地端口
	listener, err := net.Listen(Network, Address)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}
	log.Println(Address + " net.Listing...")
	// 新建gRPC服务器实例
	// 默认单次接收最大消息长度为`1024*1024*4`bytes(4M)，单次发送消息最大长度为`math.MaxInt32`bytes
	// grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(1024*1024*4), grpc.MaxSendMsgSize(math.MaxInt32))
	grpcServer := grpc.NewServer()
	// 在gRPC服务器注册我们的服务
	grpc2.RegisterServerApiServer(grpcServer, &GoServer{})

	//用服务器 Serve() 方法以及我们的端口信息区实现阻塞等待，直到进程被杀死或者 Stop() 被调用
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("grpcServer.Serve err: %v", err)
	}
}
