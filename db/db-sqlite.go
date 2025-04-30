package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	"github.com/gloopai/gloop/component"
	"github.com/gloopai/gloop/lib"
)

type DbSqlite struct {
	component.Base
	Path string // 数据库路径
	Db   *gorm.DB
}

// NewDb 创建一个新的数据库实例
func NewDbSqlite(dbpath string) *DbSqlite {
	return &DbSqlite{
		Path: dbpath,
	}
}

func (d *DbSqlite) Name() string {
	return "db"
}

// 修改 Init 方法以保存数据库连接，并提供一个方法获取连接
func (d *DbSqlite) Init() {
	lib.Log.Info("Initializing SQLite database at path:", d.Path)
	// 初始化数据库连接
	// db, err := sql.Open("sqlite", d.Path)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	db, err := gorm.Open(sqlite.Open(d.Path), &gorm.Config{})
	if err != nil {
		return
	}
	// 将数据库连接保存到结构体中
	d.Db = db
	lib.Log.Info("SQLite database initialized successfully")
}

/*  */
func (d *DbSqlite) Close() {

}

// 提供一个方法来获取数据库连接
func (d *DbSqlite) GetConnection() *gorm.DB {
	return d.Db
}
