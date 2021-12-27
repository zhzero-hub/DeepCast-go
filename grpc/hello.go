package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
)

func HelloClient() {
	ctx := context.Background()
	// 建立连接到gRPC服务
	conn, err := grpc.Dial("127.0.0.1:5002", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// 函数结束时关闭连接
	defer conn.Close()
	// 创建Waiter服务的客户端
	t := NewTrainApiClient(conn)
	tr, err := t.SayHello(ctx, &SayHelloRequest{
		Msg: "Hello world",
	})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Println(tr)
}
