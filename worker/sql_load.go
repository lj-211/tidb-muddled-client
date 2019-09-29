package worker

import (
	//"github.com/pkg/errors"
	"log"
)

func PushSqlToCoordinate(sql string) error {
	log.Print("sql语句: ", sql)
	return nil
}
