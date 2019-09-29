package coordinate

import (
	"context"

	"github.com/jinzhu/gorm"
)

// 任务协调器
//	1. 压入任务 PushTask
//	2. 查询是否完成 IsDone
//  3. 执行任务 DoTask()

type TaskProcesser func(*gorm.DB, string) error

type Coordinater interface {
	Start(context.Context, TaskProcesser) error
	PushTask(context.Context, interface{}, bool) error
	Watch(context.Context) error
	DoTask(context.Context) error
	IsDone(context.Context) bool
}

// init -> push task
// watch -> get task -> run task -> commit task -> watch ....
