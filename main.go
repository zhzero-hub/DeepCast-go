package main

import (
	rtmp "DeepCast/livego-rtmp-encrypt"
	"DeepCast/server"
	"DeepCast/train"
	"context"
	"os"
)

func main() {
	ctx := context.Background()
	c := make(chan os.Signal, 1)
	// InitLog(ctx)
	train.Init(&ctx, c)
	go server.StartGoServer(&ctx, c)
	go server.StartWebServer(&ctx, make(chan os.Signal, 1))
	go rtmp.StartRtmpServer()
	select {}
}
