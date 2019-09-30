package main

import (
	//"log"

	"github.com/pkg/errors"

	"github.com/lj-211/tidb-muddled-client/common"
)

func SqlWorker(sql string) error {
	if sql == "" {
		return common.ParamInvalidErr
	}

	//log.Println("exec: ", sql)
	err := Db.Exec(sql).Error
	if err != nil {
		return errors.Wrap(err, "执行任务sql失败")
	}

	return nil
}
