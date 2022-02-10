package train

import (
	"context"
	"testing"
)

func TestInit(t *testing.T) {
	ctx := context.Background()
	Init(&ctx)
	ChooseEdgeLocationWithKMeans(&ctx)
}
