package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"

	"github.com/gloopai/gloop/component"
)

type Db struct {
	component.Base
}

// NewDb 创建一个新的数据库实例
func NewDb() *Db {
	return &Db{}
}

func (d *Db) Name() string {
	return "db"
}

func (d *Db) Init() {
	// 初始化数据库连接
	db, err := sql.Open("sqlite", "./mydb.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		log.Fatal(err)
	}
}
