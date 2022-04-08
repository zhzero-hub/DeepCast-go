package livego

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"path"
	"runtime"
	"syscall"
	"time"

	"DeepCast/livego-rtmp-encrypt/configure"
	"DeepCast/livego-rtmp-encrypt/protocol/api"
	"DeepCast/livego-rtmp-encrypt/protocol/hls"
	"DeepCast/livego-rtmp-encrypt/protocol/httpflv"
	"DeepCast/livego-rtmp-encrypt/protocol/rtmp"

	log "github.com/sirupsen/logrus"
)

var VERSION = "master"
var LiveChan = make(chan int)

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

func startRtmp(stream *rtmp.RtmpStream, hlsServer *hls.Server) {
	rtmpAddr := configure.Config.GetString("rtmp_addr")

	rtmpListen, err := net.Listen("tcp", rtmpAddr)
	if err != nil {
		log.Fatal(err)
	}

	var rtmpServer *rtmp.Server

	if hlsServer == nil {
		rtmpServer = rtmp.NewRtmpServer(stream, nil)
		log.Info("HLS server disable....")
	} else {
		rtmpServer = rtmp.NewRtmpServer(stream, hlsServer)
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

	log.Infof(`
     _     _            ____       
    | |   (_)_   _____ / ___| ___  
    | |   | \ \ / / _ \ |  _ / _ \ 
    | |___| |\ V /  __/ |_| | (_) |
    |_____|_| \_/ \___|\____|\___/ 
        version: %s
	`, VERSION)

	apps := configure.Applications{}
	configure.Config.UnmarshalKey("server", &apps)
	for _, app := range apps {
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

		startRtmp(stream, hlsServer)
	}
}

func StartFfmpeg(resolution string) {
	//cmd := exec.Command("ffmpeg", "-i", rtmpServer, "-vcodec", "libx264", "-vprofile", "baseline", "acodec", "aac", "-strict", "-2", "-s", resolution, "-f", "flv", rtmpNginx)
	// time.Sleep(2 * time.Second)
	in := bytes.Buffer{}
	outInfo := bytes.Buffer{}
	cmd := exec.Command("bash")

	cmd.Stdout = &outInfo
	cmd.Stdin = &in

	cmdString := `ffmpeg -i rtmp://localhost:1935/live/live -vcodec libx264 -vprofile baseline -acodec aac -strict -2 -s ` + resolution + ` -f flv rtmp://localhost:1936/live/live`

	in.WriteString(cmdString)

	log.Debugf("Start ffmpeg: %s\n", resolution)
	err := cmd.Start()
	if err != nil {
		log.Warning(err.Error())
	}
	go func() {
		if err = cmd.Wait(); err != nil {
			log.Warning(err.Error())
			return
		} else {
			log.Debugln(cmd.ProcessState.Pid())
			log.Debugln(cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
			log.Debugln(outInfo.String())
		}
	}()
	select {
	case <-LiveChan:
		cmd.Process.Kill()
		return
	}
}

// ffmpeg -i rtmp://localhost:1935/live/live -vcodec libx264 -vprofile baseline -acodec aac -strict -2 -f flv rtmp://localhost:1936/live/live
