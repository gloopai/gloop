package pkg

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MysqlClientOptions struct {
	DSN string // 数据库连接字符串
}

type MysqlClient struct {
	Id   string // 数据库 ID
	DSN  string // 数据库连接字符串
	Conn *gorm.DB
}

// NewMysqlClient 创建一个新的 MySQL 客户端实例
func NewMysqlClient(opt MysqlClientOptions) *MysqlClient {
	return &MysqlClient{
		DSN: opt.DSN,
	}
}

func (d *MysqlClient) Name() string {
	return "mysql"
}

// 修改 Init 方法以保存数据库连接，并提供一个方法获取连接
func (d *MysqlClient) Init() {
	// d.printInfo()

	// 设置 gorm 的日志级别
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	// 连接 MySQL 数据库
	db, err := gorm.Open(mysql.Open(d.DSN), gormConfig)
	if err != nil {
		fmt.Printf("failed to open database: %v\n", err)
		d.Conn = nil
		return
	}

	// 将数据库连接保存到结构体中
	d.Conn = db
	// fmt.Println("MySQL database initialized successfully")
}

func (d *MysqlClient) Start() error {
	return nil
}

func (d *MysqlClient) Close() {}

func (d *MysqlClient) Destroy() {}

func (d *MysqlClient) SetEnv(env *interface{}) {}

func (d *MysqlClient) GetEnv() *interface{} {
	return nil
}

// 提供一个方法来获取数据库连接
func (d *MysqlClient) GetConnection() *gorm.DB {
	return d.Conn
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
