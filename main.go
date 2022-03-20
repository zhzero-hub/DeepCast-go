package main

import (
	"DeepCast/server"
	"DeepCast/train"
	"context"
	"os"
)

func main() {
	ctx := context.Background()
	c := make(chan os.Signal, 1)
	train.Init(&ctx, c)
	server.StartGoServer(&ctx, c)
}
