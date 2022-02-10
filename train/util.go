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

func (l *Location) DistanceCal(other *Location) float64 {
	return math.Sqrt(math.Pow(l.Lat-other.Lat, 2) + math.Pow(l.Long-other.Long, 2))
}

func (v *Viewer) LatencyCal(device DeviceCommon) {
	if strings.Contains(device.Name, "Edge") {
		v.Latency = rand.Float64()*(ViewerToCdnLatencyUpperLimit-ViewerToCdnLatencyLowerLimit) + ViewerToCdnLatencyLowerLimit
	} else if strings.Contains(device.Name, "edge") {
		if latency := v.Location.DistanceCal(&device.Location) * LatencyInGeo; latency > ViewerToEdgeLatencyUpperLimit {
			v.Latency = ViewerToEdgeLatencyUpperLimit
		} else {
			v.Latency = latency
		}
	}
}
