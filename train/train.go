package train

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

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
	}
	if err := LoadUserBandWidthDataset(*ctx); err != nil {
		log.Fatalf("加载用户位置数据失败, %v", err)
		return
	}
	GenerateRandomUserLocation(ctx)
	//log.Printf("Viewer: %v\n", (*ctx).Value("viewer"))
	//log.Printf("System: %v\n", (*ctx).Value("system"))
}

func LoadDatasetInTimeOrder(ctx *context.Context) {
	viewers := (*ctx).Value("viewer").(*ViewerInfo)
	taskManager := (*ctx).Value("taskManager").(*TaskManager)
	for _, viewer := range viewers.viewer {
		for _, liveInfo := range viewer.LiveInfo {
			taskManager.AddTask(viewer, liveInfo)
		}
	}
}

func InitSignalInterrupt(ctx *context.Context, c chan os.Signal) {
	go func() {
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGINT)
		select {
		case <-c:
			log.Println("signal received, stopping")
			file := (*ctx).Value("log")
			if file != nil {
				file.(*os.File).Close()
			}
			(*ctx).Done()
			os.Exit(0)
		}
	}()
}

func InitLog(ctx *context.Context) {
	path, _ := os.Getwd()
	path += "/output/"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
	file, err := os.OpenFile(path+"log", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("打开日志文件失败, %v\n", err)
		return
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(file)
	*ctx = context.WithValue(*ctx, "log", file)
}

func Init(ctx *context.Context, c chan os.Signal) {
	InitSignalInterrupt(ctx, c)
	InitDataset(ctx)
	InitTaskManager(ctx)
	LoadDatasetInTimeOrder(ctx)
	ChooseEdgeLocationWithKMeans(ctx)
}

//func StartTrain(ctx *context.Context) {
//	system := (*ctx).Value("system").(*System)
//	taskManager := (*ctx).Value("taskManager").(*TaskManager)
//	for {
//		if tasks := taskManager.GetTask(); tasks == nil {
//			taskManager.growth()
//		} else {
//			//for _, task := range tasks {
//			log.Println(tasks)
//			inboundUsed := make([]float64, 0)
//			outboundUsed := make([]float64, 0)
//			computeUsed := make([]float64, 0)
//			for _, inbound := range system.InboundMap {
//				inboundUsed = append(inboundUsed, *inbound)
//			}
//			for _, outbound := range system.OutboundMap {
//				outboundUsed = append(outboundUsed, *outbound)
//			}
//			for _, compute := range system.ComputationMap {
//				computeUsed = append(inboundUsed, *compute)
//			}
//			SendState(ctx, &rpc.State{
//				InboundBandwidthUsage: &rpc.InboundBandwidthUsage{
//					InboundBandwidthUsage: inboundUsed,
//				},
//				OutboundBandwidthUsage: &rpc.OutboundBandwidthUsage{
//					OutboundBandwidthUsage: outboundUsed,
//				},
//				ComputationResourceUsage: &rpc.ComputationResourceUsage{
//					ComputationResourceUsage: computeUsed,
//				},
//			})
//			//}
//			taskManager.TimeGrowth()
//		}
//	}
//}
