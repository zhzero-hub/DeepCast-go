package train

import (
	"DeepCast/grpc"
	"context"
	"log"
)

func InitDataset(ctx context.Context) {
	system := InitEdgeSystemInfo(ctx)
	ctx = context.WithValue(ctx, "system", system)
	if viewerDataset, err := LoadUserViewingDataset(ctx); err != nil {
		log.Fatalf("加载用户观看数据失败, %v", err)
		return
	} else {
		ctx = context.WithValue(ctx, "viewer", viewerDataset)
	}
	if err := LoadUserLocationDataset(ctx); err != nil {
		log.Fatalf("加载用户位置数据失败, %v", err)
		return
	} else {
		log.Printf("%v", ctx.Value("viewer"))
		log.Printf("%v", ctx.Value("system"))
	}
}

func TakeAction(ctx context.Context, action *grpc.Action) {

}
