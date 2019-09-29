package main

import (
	"github.com/jinzhu/gorm"

	"github.com/lj-211/tidb-muddled-client/coordinate"
	"github.com/lj-211/tidb-muddled-client/loader"
)

// config file path
var CfgPath string = ""

// global config obj
type DbConfig struct {
	Dsn  string
	Idle int
	Open int
}

type SqlConfig struct {
	Fpath string
}

type CliConfig struct {
	Db  DbConfig
	Cdb DbConfig
	Sql SqlConfig
}

var Config CliConfig

var Db *gorm.DB = nil

// loader
var SqlLoader loader.Loader
var TaskCoordinater coordinate.Coordinater

// 伙伴id
var Partners []string = make([]string, 0)

// 批次id
var BatchId string = ""
var Id string = ""
