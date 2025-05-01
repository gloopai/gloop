package db

import (
	"fmt"

	"github.com/gloopai/gloop/core"
	"github.com/gloopai/gloop/lib"
	components "github.com/gloopai/gloop/modules"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

type Db struct {
	components.Base
	Id   string // 数据库 ID
	Path string // 数据库路径
	Db   *gorm.DB
}

// NewDb 创建一个新的数据库实例
func NewDbSqlite(opt DbOptions) *Db {
	return &Db{
		Path: opt.DbPath,
	}
}

func (d *Db) Name() string {
	return "db"
}

// 修改 Init 方法以保存数据库连接，并提供一个方法获取连接
func (d *Db) Init() {
	d.printInfo()

	lib.Log.Info("Initializing SQLite database at path:", d.Path)
	db, err := gorm.Open(sqlite.Open(d.Path), &gorm.Config{})
	if err != nil {
		return
	}
	// 将数据库连接保存到结构体中
	d.Db = db
	lib.Log.Info("SQLite database initialized successfully")
}

/*  */
func (d *Db) Close() {}

func (d *Db) printInfo() {
	infos := make([]string, 0, 2)
	infos = append(infos, fmt.Sprintf("ID: %s", d.Id))
	infos = append(infos, fmt.Sprintf("name: %s", d.Name()))
	infos = append(infos, fmt.Sprintf("Path: %s", d.Path))
	core.PrintBoxInfo(d.Name(), infos...)
}

// 提供一个方法来获取数据库连接
func (d *Db) GetConnection() *gorm.DB {
	return d.Db
}
