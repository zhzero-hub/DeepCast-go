package main

import (
	"DeepCast/server"
	"DeepCast/train"
	"context"
)

func main() {
	ctx := context.Background()
	train.Init(&ctx)
	server.StartGoServer()
	train.StartTrain(&ctx)
}
