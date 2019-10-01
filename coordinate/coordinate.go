package coordinate

import (
	"context"
)

const (
	DoneState_Unknown = iota
	DoneState_Doing
	DoneState_ErrOccur
	DoneState_OverTime
	DoneState_OK
)

func DoneStateToStr(st int) string {
	switch {
	case st == DoneState_Unknown:
		return "未知"
	case st == DoneState_OK:
		return "已完成"
	case st == DoneState_OverTime:
		return "批次任务超时"
	case st == DoneState_ErrOccur:
		return "任务发生错误"
	default:
		return "非法的完成状态"
	}

	return ""
}

type TaskRst struct {
	DoneState int
	Msg       string
}

// 任务处理原型
type TaskProcesser func(string) error

type TaskInfo struct {
	BatchId string
	Id      string
	Sql     string
}

// 任务协调器
// init -> push task
// watch -> get task -> do task -> commit task -> watch ....
type Coordinater interface {
	// 启动协调器
	Start(context.Context, TaskProcesser) error
	// 注册任务
	// NOTE
	//	注册完成时，需要调用PushTask(ctx, nil, true)
	//	如果done == true的话，只处理推送任务完成逻辑，忽略data
	PushTask(context.Context, TaskInfo, bool) error
	// 循环监听任务调度
	Watch(context.Context) error
	// 执行任务
	DoTask(context.Context) error
	// 阻塞检查是否完成
	BlockCheckDone(context.Context) TaskRst
}
