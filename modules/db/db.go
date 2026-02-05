package db

import (
	"fmt"

	"github.com/gloopai/gloop/modules"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DbService struct {
	modules.Base
	Id  string // 数据库 ID
	DSN string // 数据库连接字符串
	Db  *gorm.DB
}

// NewDb 创建一个新的数据库实例
func NewDb(opt DbOptions) *DbService {
	return &DbService{
		DSN: opt.DSN,
	}
}

func (d *DbService) Name() string {
	return "db"
}

// 修改 Init 方法以保存数据库连接，并提供一个方法获取连接
func (d *DbService) Init() {
	d.printInfo()

	// 设置 gorm 的日志级别
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	// 连接 MySQL 数据库
	db, err := gorm.Open(mysql.Open(d.DSN), gormConfig)
	if err != nil {
		fmt.Printf("failed to open database: %v\n", err)
		d.Db = nil
		return
	}

	// 将数据库连接保存到结构体中
	d.Db = db
	fmt.Println("MySQL database initialized successfully")
}

/*  */
func (d *DbService) Close() {}

func (d *DbService) printInfo() {
	infos := make([]string, 0, 2)
	infos = append(infos, fmt.Sprintf("ID: %s", d.Id))
	infos = append(infos, fmt.Sprintf("name: %s", d.Name()))
	infos = append(infos, fmt.Sprintf("DSN: %s", d.DSN))
	infos = append(infos, "driver: MYSQL")
	// modules.PrintBoxInfo(d.Name(), infos...)
}

// 提供一个方法来获取数据库连接
func (d *DbService) GetConnection() *gorm.DB {
	return d.Db
}

/* 数据表初始化 */
func AutoMigrate(db *gorm.DB, model interface{}) error {
	if db == nil {
		return fmt.Errorf("database connection is nil, cannot migrate table")
	}
	// Automatically migrate the schema, ensuring the table structure matches the model struct
	if err := db.AutoMigrate(model); err != nil {
		return fmt.Errorf("failed to migrate table: %w", err)
	}
	return nil
}
