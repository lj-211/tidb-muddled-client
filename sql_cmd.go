package main

import (
	"log"

	"github.com/jinzhu/gorm"

	"github.com/lj-211/tidb-muddled-client/common"
)

func SqlWorker(db *gorm.DB, sql string) error {
	if db == nil {
		return common.NilInputErr
	}
	if sql == "" {
		return common.ParamInvalidErr
	}

	log.Println("exec: ", sql)

	return nil
}