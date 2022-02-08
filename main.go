package main

import (
	"DeepCast/train"
	"context"
)

func main() {
	ctx := context.Background()
	train.Init(&ctx)
	train.StartTrain(&ctx)
}
