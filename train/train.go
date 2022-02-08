package train

import (
	rpc "DeepCast/grpc"
	"context"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	PythonApiAddress = "127.0.0.1:5002"
)

var pythonServer *rpc.TrainApiClient

func InitDataset(ctx *context.Context) {
	system := InitEdgeSystemInfo(*ctx)
	*ctx = context.WithValue(*ctx, "system", system)
	if viewerDataset, err := LoadUserViewingDataset(*ctx); err != nil {
		log.Fatalf("加载用户观看数据失败, %v", err)
		return
	} else {
		*ctx = context.WithValue(*ctx, "viewer", viewerDataset)
	}
	if err := LoadUserLocationDataset(*ctx); err != nil {
		log.Fatalf("加载用户位置数据失败, %v", err)
		return
	} else {
		log.Printf("Viewer: %v\n", (*ctx).Value("viewer"))
		log.Printf("System: %v\n", (*ctx).Value("system"))
	}
}

func LoadDatasetInTimeOrder(ctx *context.Context) {
	viewers := (*ctx).Value("viewer").(*map[string]*Viewer)
	taskManager := (*ctx).Value("taskManager").(*TaskManager)
	for _, viewer := range *viewers {
		for _, liveInfo := range viewer.LiveInfo {
			taskManager.AddTask(viewer, liveInfo)
		}
	}
	log.Printf("%v", taskManager)
}

func InitGrpcClient(ctx *context.Context) {
	if conn, err := grpc.Dial(PythonApiAddress, grpc.WithInsecure()); err != nil {
		log.Fatalf("连接Python rpc服务端失败: %v\n", err)
	} else {
		// 创建Waiter服务的客户端
		t := rpc.NewTrainApiClient(conn)
		pythonServer = &t
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGINT)
			select {
			case <-c:
				log.Println("signal received, stopping")
				if conn != nil {
					err = conn.Close()
					if err != nil {
						log.Fatalf("关闭Python rpc服务端失败: %v\n", err)
						return
					}
				}
				return
			}
		}()
	}
}

func Init(ctx *context.Context) {
	InitGrpcClient(ctx)
	InitDataset(ctx)
	InitTaskManager(ctx)
	LoadDatasetInTimeOrder(ctx)
}

func StartTrain(ctx *context.Context) {
	system := (*ctx).Value("system").(*System)
	taskManager := (*ctx).Value("taskManager").(*TaskManager)
	for {
		if tasks := taskManager.GetTask(); tasks == nil || len(tasks) == 0 {
			taskManager.TimeGrowth()
		} else {
			for _, task := range tasks {
				log.Println(task)
				inboundUsed := make([]float64, 0)
				outboundUsed := make([]float64, 0)
				computeUsed := make([]float64, 0)
				for _, inbound := range system.InboundMap {
					inboundUsed = append(inboundUsed, *inbound)
				}
				for _, outbound := range system.OutboundMap {
					outboundUsed = append(outboundUsed, *outbound)
				}
				for _, compute := range system.ComputationMap {
					computeUsed = append(inboundUsed, *compute)
				}
				SendState(ctx, &rpc.State{
					InboundBandwidthUsage: &rpc.InboundBandwidthUsage{
						InboundBandwidthUsage: inboundUsed,
					},
					OutboundBandwidthUsage: &rpc.OutboundBandwidthUsage{
						OutboundBandwidthUsage: outboundUsed,
					},
					ComputationResourceUsage: &rpc.ComputationResourceUsage{
						ComputationResourceUsage: computeUsed,
					},
				})
			}
			taskManager.TimeGrowth()
		}
	}
}

func SendState(ctx *context.Context, state *rpc.State) {
	if resp, err := (*pythonServer).TrainStep(*ctx, &rpc.TrainStepRequest{
		State: state,
		Base: &rpc.Base{
			RetCode: 0,
			RetMsg:  "Success",
			Extra:   make(map[string]string, 0),
		},
	}); err != nil {
		log.Fatalf("发送状态失败: %v\n", err)
		return
	} else {
		log.Printf("%v", resp)
	}
}

func TakeAction(ctx context.Context, action *rpc.Action) {

}
