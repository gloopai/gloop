package modules

import (
	"fmt"

	"github.com/gloopai/gloop/events"
)

type ComponentEnv struct {
	Node   *Node
	Mysql  *MysqlClient
	Rdb    *Rdb
	Events *events.EventBus
}

// GetNode 获取 Node 组件实例
func (e *ComponentEnv) GetNode() (*Node, error) {
	if e.Node == nil {
		return nil, fmt.Errorf("node is not initialized")
	}
	return e.Node, nil
}

// GetMysql 获取 MysqlClient 组件实例
func (e *ComponentEnv) GetMysql() (*MysqlClient, error) {
	if e.Mysql == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	return e.Mysql, nil
}

// GetRdb 获取 Rdb 组件实例
func (e *ComponentEnv) GetRdb() (*Rdb, error) {
	if e.Rdb == nil {
		return nil, fmt.Errorf("redis is not initialized")
	}
	return e.Rdb, nil
}

// GetEvents 获取 EventBus 组件实例
func (e *ComponentEnv) GetEvents() (*events.EventBus, error) {
	if e.Events == nil {
		return nil, fmt.Errorf("event bus is not initialized")
	}
	return e.Events, nil
}
