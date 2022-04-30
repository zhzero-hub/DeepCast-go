package main

import (
	"DeepCast/livego-rtmp-encrypt"
	"github.com/go-cmd/cmd"
	"strings"
	"testing"
)

func TestStartFfmpeg(t *testing.T) {
	resolution := "1920x1080"
	cmdString := `ffmpeg -i rtmp://localhost:1935/live/live -vcodec libx264 -vprofile baseline -acodec aac -strict -2 -s ` + resolution + ` -f flv rtmp://localhost:1936/live/live`
	args := strings.Split(cmdString, " ")
	ffmpeg := cmd.NewCmd(args[0], args[1:]...)
	livego.StartFfmpeg(ffmpeg)
}
