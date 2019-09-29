package coordinate

import (
	"context"
)

// 任务处理原型
type TaskProcesser func(string) error

// 任务协调器
// init -> push task
// watch -> get task -> run task -> commit task -> watch ....
type Coordinater interface {
	// 启动协调器
	Start(context.Context, TaskProcesser) error
	// 压入任务
	PushTask(context.Context, interface{}, bool) error
	// 循环监听任务调度
	Watch(context.Context) error
	// 执行任务
	DoTask(context.Context) error
	// 检查是否完成
	IsDone(context.Context) bool
}
