package train

import (
	"container/heap"
	"context"
	"log"
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

func (t *TaskManager) GetTask() []*Task {
	return t.task[t.time]
}