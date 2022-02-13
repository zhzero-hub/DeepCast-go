package train

import (
	"context"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

func ChooseEdgeLocationWithKMeans(ctx *context.Context) {
	viewer := (*ctx).Value("viewer").(*map[string]*Viewer)
	system := (*ctx).Value("system").(*System)

	var d clusters.Observations
	for _, viewerInfo := range *viewer {
		d = append(d, clusters.Coordinates{
			viewerInfo.Location.Lat,
			viewerInfo.Location.Long,
		})
	}

	// Partition the data points into 16 clusters
	km := kmeans.New()
	if cluster, err := km.Partition(d, 10); err != nil {
		log.Fatalf("k-means生成edge location信息失败: %v\n", err)
	} else {
		for i, c := range cluster {
			system.Edge["Edge"+strconv.FormatInt(int64(i), 10)].Location = Location{
				Name: "Cluster" + strconv.FormatInt(int64(i), 10),
				Lat:  c.Center[0],
				Long: c.Center[1],
			}
			//log.Printf("Centered at x: %.2f y: %.2f\n", c.Center[0], c.Center[1])
			//log.Printf("Matching data points: %+v\n\n", c.Observations)
		}
	}
}

func (e *Edge) TranscodingLatencyCal(viewer *Viewer) float64 {
	if _, ok := e.rates[viewer.AssignInfo.ChannelId]; !ok {
		// 没有这个channelId, 把这个channel的1440 version拿进来
		v := make([]int64, 0)
		v = append(v, 1440)
		e.rates[viewer.AssignInfo.ChannelId] = &v
		e.BandWidthInfo.InBandWidthUsed += BitRateMap[1440]
		if e.BandWidthInfo.InBandWidthUsed > e.BandWidthInfo.InBandWidthLimit {
			// 应该不会触发
			log.Fatalf("超过边缘节点的带宽限制, 边缘节点名称: %s, 带宽限制: %f, 已使用: %f\n", e.Name, e.BandWidthInfo.InBandWidthLimit, e.BandWidthInfo.InBandWidthUsed)
		}
	}
	availableVersion := int64(1440)
	for _, version := range *e.rates[viewer.AssignInfo.ChannelId] {
		if version == viewer.AssignInfo.Version {
			availableVersion = version
			break
		} else if availableVersion > viewer.AssignInfo.Version && version > availableVersion {
			availableVersion = version
		}
	}
	if availableVersion == viewer.AssignInfo.Version {
		return 0
	} else if availableVersion > viewer.AssignInfo.Version {
		e.ComputationUsed += TransCodingCpuMap[viewer.AssignInfo.Version]
		e.BandWidthInfo.OutBandWidthUsed += BitRateMap[viewer.AssignInfo.Version]
		*e.rates[viewer.AssignInfo.ChannelId] = append(*e.rates[viewer.AssignInfo.ChannelId], availableVersion)
		return TransCodingTimeMap[viewer.AssignInfo.Version]
	} else {
		log.Fatalf("默认edge能够拿到所有channel的1440 version\n")
		return 0
	}
}

func (l *Location) DistanceCal(other *Location) float64 {
	return math.Sqrt(math.Pow(l.Lat-other.Lat, 2) + math.Pow(l.Long-other.Long, 2))
}

func (device *DeviceCommon) ViewerLatencyCal(v *Viewer) float64 {
	if strings.Contains(device.Name, "Cdn") {
		v.Latency = rand.Float64()*(ViewerToCdnLatencyUpperLimit-ViewerToCdnLatencyLowerLimit) + ViewerToCdnLatencyLowerLimit
	} else if strings.Contains(device.Name, "Edge") {
		if latency := v.Location.DistanceCal(&device.Location) * LatencyInGeo; latency > ViewerToEdgeLatencyUpperLimit {
			v.Latency = ViewerToEdgeLatencyUpperLimit
		} else {
			v.Latency = latency
		}
	}
	return v.Latency
}
