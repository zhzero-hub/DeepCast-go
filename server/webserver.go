package server

import (
	grpc2 "DeepCast/grpc"
	"DeepCast/train"
	"context"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
)

const (
	// WebAddress 监听地址
	WebAddress string = "0.0.0.0:5002"
)

type WebServer struct {
	grpc2.ServiceApiServer
	c  *context.Context
	ch chan os.Signal
}

func (s *WebServer) Service(ctx context.Context, request *grpc2.ServiceRequest) (*grpc2.ServiceResponse, error) {
	log.Printf("Received service")
	return &grpc2.ServiceResponse{
		Base: &grpc2.Base{},
	}, nil
}

func (s *WebServer) SystemInfo(ctx context.Context, request *grpc2.SystemInfoRequest) (*grpc2.SystemInfoResponse, error) {
	system := (*s.c).Value("system").(*train.System)
	edges := make([]*grpc2.Device, 0)
	cdns := make([]*grpc2.Device, 0)
	system.Lock()
	defer system.Unlock()
	for _, edge := range system.Edge {
		edges = append(edges, &grpc2.Device{
			Id:      edge.Id,
			Name:    edge.Name,
			CpuCore: edge.CpuCore,
			Location: &grpc2.Location{
				Latitude:  edge.Location.Lat,
				Longitude: edge.Location.Long,
			},
			BandWidthInfo: &grpc2.BandWidthInfo{
				InboundBandwidthUsage:  edge.BandWidthInfo.InBandWidthUsed,
				InboundBandwidthLimit:  edge.BandWidthInfo.InBandWidthLimit,
				OutboundBandwidthUsage: edge.BandWidthInfo.OutBandWidthUsed,
				OutboundBandwidthLimit: edge.BandWidthInfo.OutBandWidthLimit,
			},
			LatencyToUpper:   edge.LatencyToUpper,
			ComputationUsage: edge.ComputationUsed,
		})
	}
	for _, cdn := range system.Cdn {
		cdns = append(cdns, &grpc2.Device{
			Id:      cdn.Id,
			Name:    cdn.Name,
			CpuCore: cdn.CpuCore,
			Location: &grpc2.Location{
				Latitude:  cdn.Location.Lat,
				Longitude: cdn.Location.Long,
			},
			BandWidthInfo: &grpc2.BandWidthInfo{
				InboundBandwidthUsage:  cdn.BandWidthInfo.InBandWidthUsed,
				InboundBandwidthLimit:  cdn.BandWidthInfo.InBandWidthLimit,
				OutboundBandwidthUsage: cdn.BandWidthInfo.OutBandWidthUsed,
				OutboundBandwidthLimit: cdn.BandWidthInfo.OutBandWidthLimit,
			},
			LatencyToUpper:   cdn.LatencyToUpper,
			ComputationUsage: cdn.ComputationUsed,
		})
	}
	log.Printf("Return system info")
	return &grpc2.SystemInfoResponse{
		Base: &grpc2.Base{
			RetCode: 0,
			RetMsg:  "Success",
			Extra:   make(map[string]string),
		},
		SystemInfo: &grpc2.SystemInfo{
			Edges: edges,
			Cdn:   cdns,
		},
	}, nil
}

func (s *WebServer) TaskManagerInfo(ctx context.Context, request *grpc2.TaskManagerInfoRequest) (*grpc2.TaskManagerInfoResponse, error) {
	taskManager := (*s.c).Value("taskManager").(*train.TaskManager)
	taskManager.Lock()
	defer taskManager.Unlock()
	tasks := taskManager.GetAllTasks()
	solved := taskManager.GetAllSolved()
	userInfos := make([]*grpc2.UserInfo, 0)
	solves := make([]*grpc2.Solve, 0)
	max, _ := strconv.ParseInt(request.Base.Extra["max"], 10, 64)
	count := int64(0)
	for _, task := range tasks {
		info := task.GetTask()
		userInfos = append(userInfos, &grpc2.UserInfo{
			Location: &grpc2.Location{
				Latitude:  info.Location.Lat,
				Longitude: info.Location.Long,
			},
			ChannelId: info.ChannelId,
			Version:   info.VersionBit,
			UserId:    info.Viewer.Id,
		})
		if count++; count >= max {
			break
		}
	}
	count = 0
	for viewerWithWatchChannel, deviceName := range solved {
		solves = append(solves, &grpc2.Solve{
			UserInfo: &grpc2.UserInfo{
				Location: &grpc2.Location{
					Latitude:  viewerWithWatchChannel.Location.Lat,
					Longitude: viewerWithWatchChannel.Location.Long,
				},
				ChannelId: viewerWithWatchChannel.ChannelId,
				Version:   viewerWithWatchChannel.VersionBit,
				UserId:    viewerWithWatchChannel.Viewer.Id,
			},
			DeviceName: deviceName,
		})
		if count++; count >= max {
			break
		}
	}
	log.Printf("Return task manager")
	return &grpc2.TaskManagerInfoResponse{
		Base: &grpc2.Base{
			RetCode: 0,
			RetMsg:  "Success",
			Extra:   make(map[string]string),
		},
		TaskManagerInfo: &grpc2.TaskManagerInfo{
			Time:     taskManager.GetTime(),
			UserInfo: userInfos,
			Solved:   solves,
		},
	}, nil
}

func (s *WebServer) BackgroundInfo(ctx context.Context, request *grpc2.BackgroundInfoRequest) (*grpc2.BackgroundInfoResponse, error) {
	taskManager := (*s.c).Value("taskManager").(*train.TaskManager)
	log.Printf("Receive background info")
	var location *grpc2.Location
	if _, ok := request.Base.Extra["location"]; ok {
		locations := (*s.c).Value("locations").([]train.Location)
		l := locations[rand.Intn(len(locations))]
		location = &grpc2.Location{
			Latitude:  l.Lat,
			Longitude: l.Long,
		}
	}
	return &grpc2.BackgroundInfoResponse{
		Base: &grpc2.Base{
			RetCode: 0,
			RetMsg:  "Success",
			Extra:   make(map[string]string),
		},
		BackgroundInfo: &grpc2.BackgroundInfo{
			Time:     taskManager.GetTime(),
			MaxTime:  taskManager.GetMaxTime(),
			Location: location,
		},
	}, nil
}

func StartWebServer(ctx *context.Context, c chan os.Signal) {
	// 监听本地端口
	listener, err := net.Listen(Network, WebAddress)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}
	log.Println(WebAddress + " web server net.Listing...")
	// 新建gRPC服务器实例
	// 默认单次接收最大消息长度为`1024*1024*4`bytes(4M)，单次发送消息最大长度为`math.MaxInt32`bytes
	// grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(1024*1024*4), grpc.MaxSendMsgSize(math.MaxInt32))
	grpcServer := grpc.NewServer()
	// 在gRPC服务器注册我们的服务
	grpc2.RegisterServiceApiServer(grpcServer, &WebServer{
		c:  ctx,
		ch: c,
	})

	//用服务器 Serve() 方法以及我们的端口信息区实现阻塞等待，直到进程被杀死或者 Stop() 被调用
	if err = grpcServer.Serve(listener); err != nil {
		log.Fatalf("grpcServer.Serve err: %v", err)
	}
}
