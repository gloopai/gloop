package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"

	"github.com/gloopai/gloop/component"
	"github.com/gloopai/gloop/lib"
)

type DbSqlite struct {
	component.Base
	Path string // 数据库路径
	Db   *sql.DB
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
	db, err := sql.Open("sqlite", d.Path)
	if err != nil {
		log.Fatal(err)
	}
	// 将数据库连接保存到结构体中
	d.Db = db
	lib.Log.Info("SQLite database initialized successfully")
}

/*  */
func (d *DbSqlite) Close() {
	d.Db.Close()
}

// 提供一个方法来获取数据库连接
func (d *DbSqlite) GetConnection() *sql.DB {
	return d.Db
}
