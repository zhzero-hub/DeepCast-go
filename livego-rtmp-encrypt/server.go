package livego

import (
	"fmt"
	"net"
	"path"
	"runtime"
	"time"

	"DeepCast/livego-rtmp-encrypt/configure"
	"DeepCast/livego-rtmp-encrypt/protocol/api"
	"DeepCast/livego-rtmp-encrypt/protocol/hls"
	"DeepCast/livego-rtmp-encrypt/protocol/httpflv"
	"DeepCast/livego-rtmp-encrypt/protocol/rtmp"

	"github.com/go-cmd/cmd"

	log "github.com/sirupsen/logrus"
)

var VERSION = "master"

func startHls() *hls.Server {
	hlsAddr := configure.Config.GetString("hls_addr")
	hlsListen, err := net.Listen("tcp", hlsAddr)
	if err != nil {
		log.Fatal(err)
	}

	hlsServer := hls.NewServer()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("HLS server panic: ", r)
			}
		}()
		log.Info("HLS listen On ", hlsAddr)
		hlsServer.Serve(hlsListen)
	}()
	return hlsServer
}

func startRtmp(stream *rtmp.RtmpStream, hlsServer *hls.Server, encrypt bool) {
	rtmpAddr := configure.Config.GetString("rtmp_addr")

	rtmpListen, err := net.Listen("tcp", rtmpAddr)
	if err != nil {
		log.Fatal(err)
	}

	var rtmpServer *rtmp.Server

	if hlsServer == nil {
		rtmpServer = rtmp.NewRtmpServer(stream, nil, encrypt)
		log.Info("HLS server disable....")
	} else {
		rtmpServer = rtmp.NewRtmpServer(stream, hlsServer, encrypt)
		log.Info("HLS server enable....")
	}

	defer func() {
		if r := recover(); r != nil {
			log.Error("RTMP server panic: ", r)
		}
	}()
	log.Info("RTMP Listen On ", rtmpAddr)
	rtmpServer.Serve(rtmpListen)
}

func startHTTPFlv(stream *rtmp.RtmpStream) {
	httpflvAddr := configure.Config.GetString("httpflv_addr")

	flvListen, err := net.Listen("tcp", httpflvAddr)
	if err != nil {
		log.Fatal(err)
	}

	hdlServer := httpflv.NewServer(stream)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("HTTP-FLV server panic: ", r)
			}
		}()
		log.Info("HTTP-FLV listen On ", httpflvAddr)
		hdlServer.Serve(flvListen)
	}()
}

func startAPI(stream *rtmp.RtmpStream) {
	apiAddr := configure.Config.GetString("api_addr")
	rtmpAddr := configure.Config.GetString("rtmp_addr")

	if apiAddr != "" {
		opListen, err := net.Listen("tcp", apiAddr)
		if err != nil {
			log.Fatal(err)
		}
		opServer := api.NewServer(stream, rtmpAddr)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error("HTTP-API server panic: ", r)
				}
			}()
			log.Info("HTTP-API listen On ", apiAddr)
			opServer.Serve(opListen)
		}()
	}
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf(" %s:%d", filename, f.Line)
		},
	})
}

func StartRtmpServer() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("livego panic: ", r)
			time.Sleep(1 * time.Second)
		}
	}()
	apps := configure.Applications{}
	configure.Config.UnmarshalKey("server", &apps)
	for _, app := range apps {
		log.Info("Encryption: ", app.Encrypt)
		stream := rtmp.NewRtmpStream()
		if msg, err := configure.RoomKeys.GetKey("live"); err != nil {
			log.Error("get room key error: ", err)
		} else {
			log.Printf("room key: %s\n", msg)
		}

		var hlsServer *hls.Server
		if app.Hls {
			hlsServer = startHls()
		}
		if app.Flv {
			startHTTPFlv(stream)
		}
		if app.Api {
			startAPI(stream)
		}

		startRtmp(stream, hlsServer, app.Encrypt)
	}
}

func StartFfmpeg(ffmpeg *cmd.Cmd) {
	//cmd := exec.Command("ffmpeg", "-i", rtmpServer, "-vcodec", "libx264", "-vprofile", "baseline", "acodec", "aac", "-strict", "-2", "-s", resolution, "-f", "flv", rtmpNginx)
	// time.Sleep(2 * time.Second)

	statusChan := ffmpeg.Start()
	if statusChan == nil {
		log.Warning("Error start ffmpeg\n")
	}
	ticker := time.NewTicker(2 * time.Second)
	// Print last line of stdout every 2s
	go func() {
		for range ticker.C {
			status := ffmpeg.Status()
			n := len(status.Stdout)
			if n > 0 {
				log.Println(status.Stdout[n-1])
			}
		}
	}()
	select {
	case finalStatus := <-statusChan:
		// done
		log.Println(finalStatus)
	}
}

// ffmpeg -i rtmp://localhost:1935/live/live -vcodec libx264 -vprofile baseline -acodec aac -strict -2 -f flv rtmp://localhost:1936/live/live
