package main

import (
	"context"
	"log"

	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/lj-211/tidb-muddled-client/coordinate"
	"github.com/lj-211/tidb-muddled-client/loader"
)

func checkGlobal() error {
	if CfgPath == "" {
		return errors.New("配置文件路径未设置")
	}

	if len(Partners) == 0 {
		return errors.New("没有协调同步的客户端")
	}

	if BatchId == "" {
		return errors.New("没有批次id,无法进行任务协调")
	}

	return nil
}

func initConfig() error {
	var err error = nil

	if _, terr := toml.DecodeFile(CfgPath, &Config); err != nil {
		err = errors.Wrap(terr, "toml解码失败")
	}
	log.Printf("加载配置: %+v\n", Config)

	return err
}

func initDbConn() error {
	db, err := gorm.Open("mysql", Config.Db.Dsn)
	if err != nil {
		return errors.Wrap(err, "db连接异常")
	}

	db.DB().SetMaxIdleConns(Config.Db.Idle)
	db.DB().SetMaxOpenConns(Config.Db.Open)

	Db = db

	return nil
}

func initCoordinater() error {
	db, err := gorm.Open("mysql", Config.Cdb.Dsn)
	if err != nil {
		return errors.Wrap(err, "初始化协调器db失败")
	}

	db.DB().SetMaxIdleConns(Config.Cdb.Idle)
	db.DB().SetMaxOpenConns(Config.Cdb.Open)

	dbc, err := coordinate.NewDbCoordinate(Id, BatchId, Partners, db)
	if err != nil {
		return errors.Wrap(err, "初始化协调器失败")
	}

	db.LogMode(false)

	TaskCoordinater = dbc

	TaskCoordinater.Start(context.Background(), SqlWorker)

	return nil
}

func initLoader() error {
	SqlLoader = &loader.FileSqlLoader{}
	return nil
}

func loadData() error {
	err := SqlLoader.Load(Config.Sql.Fpath, PushSqlToCoordinate)
	if err != nil {
		return errors.Wrap(err, "加载sql失败")
	}

	return nil
}
