package train

import (
	rpc "DeepCast/grpc"
	"container/heap"
	"context"
	"log"
	"math"
	"strconv"
	"strings"
)

type ViewerWithWatchChannel struct {
	*Viewer
	*LiveInfo
}

type ViewerHeap []*ViewerWithWatchChannel

func (h ViewerHeap) Len() int {
	return len(h)
}

func (h ViewerHeap) Less(i, j int) bool {
	return h[i].EndTime < h[j].EndTime
}

func (h ViewerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *ViewerHeap) Push(x interface{}) {
	*h = append(*h, x.(*ViewerWithWatchChannel))
}

func (h *ViewerHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *ViewerHeap) Top() *ViewerWithWatchChannel {
	return (*h)[0]
}

type Task struct {
	watchInfo *ViewerWithWatchChannel
}

type TaskManager struct {
	ctx        *context.Context
	time       int64
	task       map[int64][]*Task
	viewerList ViewerHeap
	solved     map[string]*DeviceCommon
	maxTime    int64
}

func InitTaskManager(ctx *context.Context) {
	viewerWithWatchChannel := make(ViewerHeap, 0)
	heap.Init(&viewerWithWatchChannel)
	taskManager := TaskManager{
		ctx:        ctx,
		time:       0,
		task:       make(map[int64][]*Task),
		viewerList: viewerWithWatchChannel,
	}
	*ctx = context.WithValue(*ctx, "taskManager", &taskManager)
}

func (t *TaskManager) AddTask(viewer *Viewer, liveInfo *LiveInfo) {
	startTime := liveInfo.StartTime
	if startTime == 5984 {
		log.Printf("%v", liveInfo)
	}
	if _, ok := t.task[startTime]; !ok {
		task := make([]*Task, 0)
		task = append(task, &Task{
			watchInfo: &ViewerWithWatchChannel{
				Viewer:   viewer,
				LiveInfo: liveInfo,
			},
		})
		t.task[startTime] = task
	} else {
		task := t.task[startTime]
		task = append(task, &Task{
			watchInfo: &ViewerWithWatchChannel{
				Viewer:   viewer,
				LiveInfo: liveInfo,
			},
		})
		t.task[startTime] = task
	}
}

func (t *TaskManager) RefreshTasks() {
	if len(t.viewerList) == 0 {
		return
	} else {
		for top := t.viewerList.Top(); top.EndTime <= t.time; top = t.viewerList.Top() {
			t.viewerList.Pop()
			if len(t.viewerList) == 0 {
				return
			}
		}
	}
}

func (t *TaskManager) TimeGrowth() {
	t.time++
	t.RefreshTasks()
	for _, task := range t.task[t.time] {
		t.viewerList.Push(task.watchInfo)
	}
}

func (t *TaskManager) GetTask() *Task {
	return &Task{
		watchInfo: t.viewerList.Top(),
	}
}

func (t *TaskManager) TakeAction(ctx *context.Context, action *rpc.Action) {
	// 从Base.Extra拿了什么: version, deviceId
	system := (*t.ctx).Value("system").(*System)
	viewerId := action.GetViewerId()
	viewerMap := (*t.ctx).Value("viewer").(*map[string]*Viewer)
	viewer := (*viewerMap)[viewerId]
	version, _ := strconv.ParseInt(action.Base.Extra["version"], 10, 64)
	viewer.AssignInfo = AssignInfo{
		DeviceId: action.Base.Extra["deviceId"],
		Version:  version,
	}
	if edge, ok := system.Edge["Edge"+strconv.FormatInt(action.GetAction(), 10)]; ok {
		if outBandUsage := edge.BandWidthInfo.OutBandWidthUsed + viewer.DownThroughput; outBandUsage < edge.BandWidthInfo.OutBandWidthLimit {
			edge.BandWidthInfo.OutBandWidthUsed = outBandUsage
			viewer.AssignInfo.DeviceId = edge.Name
			t.solved[viewerId] = &edge.DeviceCommon
		} else if edge.BandWidthInfo.OutBandWidthUsed+BitRateMap[240] > edge.BandWidthInfo.OutBandWidthLimit {
			// TODO: 无论如何都不够了，需要redirect
		} else {
			// 尽可能给
			var realRate int64
			for index, rate := range BitRateMap {
				if edge.BandWidthInfo.OutBandWidthUsed+rate < edge.BandWidthInfo.OutBandWidthLimit {
					if realRate < index {
						realRate = index
					}
				}
			}
			edge.BandWidthInfo.OutBandWidthUsed += BitRateMap[realRate]
			viewer.AssignInfo.DeviceId = edge.Name
			viewer.AssignInfo.Version = realRate
			t.solved[viewerId] = &edge.DeviceCommon
		}
	} else if cdn, ok := system.Cdn["Cdn"+strconv.FormatInt(action.GetAction(), 10)]; ok {
		viewer.AssignInfo.DeviceId = cdn.Name
		t.solved[viewerId] = &cdn.DeviceCommon
	} else {
		log.Fatalf("不存在的action: %v\n", action)
	}
	t.viewerList.Pop()

	reward := GetReward(ctx, viewer, action)
}

func GetReward(ctx *context.Context, viewer *Viewer, action *rpc.Action) float64 {
	var rewarda, rewardb float64
	rewarda += float64(action.GetQoePreference().Alpha1) * GetStreamingDelay(ctx, viewer, viewer.AssignInfo.DeviceId)
	rewarda += float64(action.GetQoePreference().Alpha2) * GetChannelSwitchingDelay(ctx, viewer)
	rewarda += float64(action.GetQoePreference().Alpha3) * GetMismatchLevel(ctx, viewer, action)
	rewarda *= Alpha

	rewardb += GetComputationCost(ctx, viewer)
	rewardb *= Beta
	return rewarda + rewardb
}

func GetStreamingDelay(ctx *context.Context, viewer *Viewer, deviceId string) float64 {
	var streamingDelay float64
	system := (*ctx).Value("system").(*System)
	if edge, ok := system.Edge["Edge"+deviceId]; ok {
		streamingDelay += edge.ViewerLatencyCal(viewer)
		streamingDelay += edge.TranscodingLatencyCal(viewer)
		streamingDelay += edge.LatencyToUpper
	} else if cdn, ok := system.Cdn[deviceId]; ok {
		streamingDelay += cdn.ViewerLatencyCal(viewer)
	} else {
		log.Fatalf("不存在的edge: %v\n", deviceId)
	}
	return streamingDelay
}

func GetChannelSwitchingDelay(ctx *context.Context, viewer *Viewer) float64 {
	return viewer.Latency
}

func GetMismatchLevel(ctx *context.Context, viewer *Viewer, action *rpc.Action) float64 {
	var mismatchLevel float64
	version, _ := strconv.ParseInt(action.Base.Extra["version"], 10, 64)
	mismatchLevel += math.Log(float64(version)) - math.Log(float64(viewer.AssignInfo.Version))
	return mismatchLevel
}

func GetComputationCost(ctx *context.Context, viewer *Viewer) float64 {
	if strings.Contains(viewer.AssignInfo.DeviceId, "Edge") {
		return TransCodingCpuMap[viewer.AssignInfo.Version] * Price
	} else {
		return 0
	}
}

func GetBandwidthCost(ctx *context.Context) float64 {

}
