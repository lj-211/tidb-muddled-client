package main

import (
	"context"
	//"log"

	"github.com/lj-211/tidb-muddled-client/coordinate"
)

func PushSqlToCoordinate(sql string, isLast bool) error {
	ci := coordinate.TaskInfo{
		BatchId: BatchId,
		Id:      Id,
		Sql:     sql,
	}
	return TaskCoordinater.PushTask(context.Background(), ci, isLast)
}
