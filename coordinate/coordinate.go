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
	// 注册任务
	// NOTE
	//	注册完成时，需要调用PushTask(ctx, nil, true)
	//	如果done == true的话，只处理推送任务完成逻辑，忽略data
	PushTask(context.Context, interface{}, bool) error
	// 循环监听任务调度
	Watch(context.Context) error
	// 执行任务
	DoTask(context.Context) error
	// 检查是否完成
	IsDone(context.Context) bool
}
