package modules

import "github.com/gloopai/gloop/events"

type ComponentEnv struct {
	Node   *Node
	DB     *DbService
	Events *events.EventBus
}
