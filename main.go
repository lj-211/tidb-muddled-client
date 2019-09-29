package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/lj-211/tidb-muddled-client/worker"
)

func argParse() {
	flag.StringVar(&CfgPath, "c", "", "配置文件文件路径")
	partners := ""
	flag.StringVar(&partners, "p", "", "伙伴id,exp: 1,2,3")
	flag.StringVar(&BatchId, "b", "", "同一批次客户端的批次id")
	flag.StringVar(&Id, "id", "", "节点id")

	flag.Parse()

	if partners != "" {
		Partners = strings.Split(partners, ",")
	}
}

func boot() error {
	argParse()

	var err error
	for ok := true; ok; ok = false {
		cerr := checkGlobal()
		if cerr != nil {
			err = errors.Wrap(cerr, "全局参数校验未通过")
			break
		}

		cerr = initConfig()
		if cerr != nil {
			err = errors.Wrap(cerr, "初始化配置出错")
			break
		}

		cerr = initDbConn()
		if cerr != nil {
			err = errors.Wrap(cerr, "初始化db连接失败")
			break
		}

		cerr = initCoordinater()
		if cerr != nil {
			err = errors.Wrap(cerr, "初始化协调器失败")
			break
		}

		cerr = initLoader()
		if cerr != nil {
			err = errors.Wrap(cerr, "初始化加载器失败")
			break
		}
	}

	return err
}

func StartWork() {
	err := loadData()
	if err != nil {
		log.Println("ERROR: 加载数据失败 ", err.Error())
		return
	}
	TaskCoordinater.Start(context.Background(), worker.SqlWorker)

	for {
		done := TaskCoordinater.IsDone(context.TODO())
		if done {
			break
		}
	}

	return
}

func main() {
	err := boot()
	if err != nil {
		log.Println("启动失败: ", err.Error())
		os.Exit(-1)
	}

	StartWork()

	return
}
