package component

import (
	"fmt"

	"github.com/gloopai/gloop/events"
)

type ComponentEnv struct {
	// Node   *node.Node
	Events *events.EventBus
}

// // GetNode 获取 Node 组件实例
// func (e *ComponentEnv) GetNode() (*node.Node, error) {
// 	if e.Node == nil {
// 		return nil, fmt.Errorf("node is not initialized")
// 	}
// 	return e.Node, nil
// }

// // GetMysql 获取 MysqlClient 组件实例
// func (e *ComponentEnv) GetMysql() (*pkg.MysqlClient, error) {
// 	if e.Node == nil {
// 		return nil, fmt.Errorf("node is not initialized")
// 	}
// 	return e.Node.GetMysql()
// }

// // GetRdb 获取 Rdb 组件实例
// func (e *ComponentEnv) GetRdb() (*pkg.RedisClient, error) {
// 	if e.Node == nil {
// 		return nil, fmt.Errorf("node is not initialized")
// 	}
// 	return e.Node.GetRdb()
// }

// GetEvents 获取 EventBus 组件实例
func (e *ComponentEnv) GetEvents() (*events.EventBus, error) {
	if e.Events == nil {
		return nil, fmt.Errorf("event bus is not initialized")
	}
	return e.Events, nil
}
