package modules

import "github.com/gloopai/gloop/events"

type ComponentContext struct {
	Node   *Node
	DB     *DbService
	Events *events.EventBus
}
