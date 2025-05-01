package db

type DbOptions struct {
	DbPath string `json:"db_path"` // 数据库路径
	Name   string `json:"name"`    // 数据库名称
}
