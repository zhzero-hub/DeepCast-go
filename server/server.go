package server

import (
	grpc2 "DeepCast/grpc"
	"DeepCast/train"
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"strconv"
)

const (
	// Address 监听地址
	Address string = "0.0.0.0:5001"
	// Network 网络通信协议
	Network string = "tcp"
)

type GoServer struct {
	grpc2.TrainApiServer
	c  *context.Context
	ch chan os.Signal
}

func (g *GoServer) SayHello(ctx context.Context, request *grpc2.SayHelloRequest) (*grpc2.SayHelloResponse, error) {
	log.Printf("Received hello message: %v", request.Msg)
	if request.Msg == "Over" {
		go func() {
			g.ch <- os.Interrupt
		}()
	}
	return &grpc2.SayHelloResponse{
		Msg: request.Msg,
	}, nil
}

func (g *GoServer) ResetEnv(ctx context.Context, request *grpc2.ResetEnvRequest) (*grpc2.ResetEnvResponse, error) {
	log.Printf("Received reset env")
	taskManager := (*g.c).Value("taskManager").(*train.TaskManager)
	nextState := taskManager.NextState(g.c)
	return &grpc2.ResetEnvResponse{
		Base: &grpc2.Base{
			RetCode: int64(0),
			RetMsg:  "Success",
			Extra: map[string]string{
				"channelId": nextState.UserInfo.ChannelId,
				"version":   strconv.FormatInt(nextState.UserInfo.Version, 10),
				"E":         strconv.Itoa(train.EdgeNumber),
				"channel":   strconv.Itoa(train.Channels),
				"versions":  strconv.Itoa(len(train.BitRateMap)),
			},
		},
		State: nextState,
	}, nil
}

func (g *GoServer) TrainStep(ctx context.Context, request *grpc2.TrainStepRequest) (*grpc2.TrainStepResponse, error) {
	log.Printf("Received train step")
	action := request.Action
	if request.Base == nil || action == nil {
		log.Printf("request.Base or request.Action is nil\n")
		log.Printf("req: %v\n", request)
		return &grpc2.TrainStepResponse{
			Base: &grpc2.Base{
				RetCode: int64(1),
				RetMsg:  "request.Base or request.Action is nil",
				Extra:   map[string]string{},
			},
		}, nil
	}
	log.Printf("action: %v\n", action)
	taskManager := (*g.c).Value("taskManager").(*train.TaskManager)
	reward := taskManager.TakeAction(g.c, request)
	nextState := taskManager.NextState(g.c)
	if nextState == nil {
		return &grpc2.TrainStepResponse{
			Base: &grpc2.Base{
				RetCode: int64(1),
				RetMsg:  "nextState is nil",
				Extra:   map[string]string{},
			},
			State: nil,
			Feedback: &grpc2.Feedback{
				Reward: reward,
			},
		}, nil
	}
	return &grpc2.TrainStepResponse{
		Base: &grpc2.Base{
			RetCode: int64(1),
			RetMsg:  "Success",
			Extra: map[string]string{
				"channelId": nextState.UserInfo.ChannelId,
				"version":   strconv.FormatInt(nextState.UserInfo.Version, 10),
				"E":         strconv.Itoa(train.EdgeNumber),
				"channel":   strconv.Itoa(train.Channels),
				"versions":  strconv.Itoa(len(train.BitRateMap)),
			},
		},
		State: nil,
		Feedback: &grpc2.Feedback{
			Reward: reward,
		},
	}, nil
}

func StartGoServer(ctx *context.Context, c chan os.Signal) {
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
	grpc2.RegisterTrainApiServer(grpcServer, &GoServer{
		c:  ctx,
		ch: c,
	})

	//用服务器 Serve() 方法以及我们的端口信息区实现阻塞等待，直到进程被杀死或者 Stop() 被调用
	if err = grpcServer.Serve(listener); err != nil {
		log.Fatalf("grpcServer.Serve err: %v", err)
	}
}
