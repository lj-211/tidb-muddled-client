package main

import (
	"context"
	//"log"

	"github.com/lj-211/tidb-muddled-client/coordinate"
)

func PushSqlToCoordinate(sql string, isLast bool) error {
	ci := coordinate.CmdInfo{
		BatchId: BatchId,
		NodeId:  Id,
		Sql:     sql,
	}
	return TaskCoordinater.PushTask(context.TODO(), ci, isLast)
}
