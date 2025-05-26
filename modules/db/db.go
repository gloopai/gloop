package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gloopai/gloop/modules"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

type DbService struct {
	modules.Base
	Id   string // 数据库 ID
	Path string // 数据库路径
	Db   *gorm.DB
}

// NewDb 创建一个新的数据库实例
func NewDb(opt DbOptions) *DbService {
	return &DbService{
		Path: opt.DbPath,
	}
}

func (d *DbService) Name() string {
	return "db"
}

// 修改 Init 方法以保存数据库连接，并提供一个方法获取连接
func (d *DbService) Init() {
	d.printInfo()

	// 检查数据库文件夹是否存在，不存在则创建
	dir := filepath.Dir(d.Path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("failed to create database directory: %v\n", err)
			return
		}
	}

	// lib.Log.Info("Initializing SQLite database at path:", d.Path)
	// 设置 gorm 的日志级别
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}
	db, err := gorm.Open(sqlite.Open(d.Path), gormConfig)
	if err != nil {
		return
	}
	// 将数据库连接保存到结构体中
	d.Db = db
	// lib.Log.Info("SQLite database initialized successfully")
}

/*  */
func (d *DbService) Close() {}

func (d *DbService) printInfo() {
	infos := make([]string, 0, 2)
	infos = append(infos, fmt.Sprintf("ID: %s", d.Id))
	infos = append(infos, fmt.Sprintf("name: %s", d.Name()))
	infos = append(infos, fmt.Sprintf("Path: %s", d.Path))
	infos = append(infos, "driver: SQLITE")
	modules.PrintBoxInfo(d.Name(), infos...)
}

// 提供一个方法来获取数据库连接
func (d *DbService) GetConnection() *gorm.DB {
	return d.Db
}

/* 数据表初始化 */
func AutoMigrate(db *gorm.DB, model interface{}) error {
	// Automatically migrate the schema, ensuring the table structure matches the model struct
	if err := db.AutoMigrate(model); err != nil {
		return fmt.Errorf("failed to migrate table: %w", err)
	}
	return nil
}
