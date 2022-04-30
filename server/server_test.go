package server

import (
	"DeepCast/livego-rtmp-encrypt"
	"github.com/go-cmd/cmd"
	"testing"
)

func TestStartGoServer(t *testing.T) {
	resolution := "1920x1080"
	cmdString := `ffmpeg -i rtmp://localhost:1935/live/live -vcodec libx264 -vprofile baseline -acodec aac -strict -2 -s ` + resolution + ` -f flv rtmp://localhost:1936/live/live`
	ffmpeg := cmd.NewCmd(cmdString)
	livego.StartFfmpeg(ffmpeg)
}
