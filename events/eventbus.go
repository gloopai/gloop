package events

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/gloopai/gloop/lib"
	"github.com/google/uuid"
)

// EventMessage 定义事件消息结构体
// 包含消息编号、时间戳和数据
// EventMessage 事件消息结构体
type EventMessage struct {
	ID        string      // 消息编号
	Timestamp time.Time   // 消息时间
	Data      interface{} // 消息数据
}

// Data 反序列化
func (e *EventMessage) Unmarshal(v interface{}) error {
	return lib.Convert.InterfaceToStruct(e.Data, &v)
}

// EventHandler defines the function signature for event handlers.
type EventHandler func(msg *EventMessage)

// EventHandlerWithContext defines the function signature for event handlers with context and generic type.
type EventHandlerWithContext[T any] func(ctx context.Context, data T)

// LoggerHook is a function for logging event bus actions.
type LoggerHook func(event string, action string, info map[string]interface{})

// EventBus manages event subscriptions and publishing.
type EventBus struct {
	listeners       map[string][]EventHandler
	patternHandlers []patternHandler
	lock            sync.RWMutex
}

// GenericEventBus 支持泛型、超时、日志钩子的事件总线。
type GenericEventBus[T any] struct {
	listeners       map[string][]EventHandlerWithContext[T]
	patternHandlers []struct {
		pattern string
		handler EventHandlerWithContext[T]
	}
	lock   sync.RWMutex
	logger LoggerHook
}

// patternHandler is used for storing pattern-based subscriptions.
type patternHandler struct {
	pattern string
	handler EventHandler
}

// NewEventBus creates a new EventBus instance.
func NewEventBus() *EventBus {
	return &EventBus{
		listeners:       make(map[string][]EventHandler),
		patternHandlers: []patternHandler{},
	}
}

// NewGenericEventBus 创建泛型事件总线。
func NewGenericEventBus[T any](logger LoggerHook) *GenericEventBus[T] {
	return &GenericEventBus[T]{
		listeners: make(map[string][]EventHandlerWithContext[T]),
		patternHandlers: []struct {
			pattern string
			handler EventHandlerWithContext[T]
		}{},
		logger: logger,
	}
}

// Subscribe adds a handler for a specific event.
// If event is empty or handler is nil, it does nothing.
// If the handler is already subscribed to the event, it will not be added again.
func (eb *EventBus) Subscribe(event string, handler EventHandler) {
	// lib.Log.Debugf("Subscribe event: %s, handler: %v", event, handler)
	if event == "" || handler == nil {
		return
	}
	ptr := getFuncPointer(handler)
	eb.lock.Lock()
	defer eb.lock.Unlock()
	for _, h := range eb.listeners[event] {
		if getFuncPointer(h) == ptr {
			return // already subscribed
		}
	}
	eb.listeners[event] = append(eb.listeners[event], handler)
}

// Subscribe 订阅事件。
func (eb *GenericEventBus[T]) Subscribe(event string, handler EventHandlerWithContext[T]) {
	// lib.Log.Debugf("Subscribe event: %s, handler: %v", event, handler)
	if event == "" || handler == nil {
		return
	}
	eb.lock.Lock()
	defer eb.lock.Unlock()
	for _, h := range eb.listeners[event] {
		if reflect.ValueOf(h).Pointer() == reflect.ValueOf(handler).Pointer() {
			return
		}
	}
	eb.listeners[event] = append(eb.listeners[event], handler)
	if eb.logger != nil {
		eb.logger(event, "subscribe", nil)
	}
}

// Unsubscribe removes a handler for a specific event.
// If event is empty or handler is nil, it does nothing
func (eb *EventBus) Unsubscribe(event string, handler EventHandler) {
	// lib.Log.Debugf("Unsubscribe event: %s, handler: %v", event, handler)
	if event == "" || handler == nil {
		return
	}
	eb.lock.Lock()
	defer eb.lock.Unlock()
	handlers := eb.listeners[event]
	for i, h := range handlers {
		if getFuncPointer(h) == getFuncPointer(handler) {
			eb.listeners[event] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
	if len(eb.listeners[event]) == 0 {
		delete(eb.listeners, event)
	}
}

// Once subscribes a handler that will be called only once for the event.
func (eb *EventBus) Once(event string, handler EventHandler) {
	// lib.Log.Debugf("Once event: %s, handler: %v", event, handler)
	var wrapper EventHandler
	wrapper = func(msg *EventMessage) {
		eb.Unsubscribe(event, wrapper)
		handler(msg)
	}
	eb.Subscribe(event, wrapper)
}

// SubscribePattern adds a handler for events matching a pattern (e.g. "user.*").
// 支持简单的前缀通配符（如 "user.*" 匹配所有以 "user." 开头的事件）。
func (eb *EventBus) SubscribePattern(pattern string, handler EventHandler) {
	// lib.Log.Debugf("SubscribePattern pattern: %s, handler: %v", pattern, handler)
	if pattern == "" || handler == nil {
		return
	}
	eb.lock.Lock()
	defer eb.lock.Unlock()
	for _, ph := range eb.patternHandlers {
		if ph.pattern == pattern && getFuncPointer(ph.handler) == getFuncPointer(handler) {
			return // already subscribed
		}
	}
	eb.patternHandlers = append(eb.patternHandlers, patternHandler{pattern, handler})
}

// UnsubscribePattern removes a pattern handler.
func (eb *EventBus) UnsubscribePattern(pattern string, handler EventHandler) {
	// lib.Log.Debugf("UnsubscribePattern pattern: %s, handler: %v", pattern, handler)
	if pattern == "" || handler == nil {
		return
	}
	eb.lock.Lock()
	defer eb.lock.Unlock()
	for i, ph := range eb.patternHandlers {
		if ph.pattern == pattern && getFuncPointer(ph.handler) == getFuncPointer(handler) {
			eb.patternHandlers = append(eb.patternHandlers[:i], eb.patternHandlers[i+1:]...)
			break
		}
	}
}

// Publish triggers all handlers subscribed to an event.
func (eb *EventBus) Publish(event string, data interface{}) {
	// lib.Log.Debugf("Publish event: %s, data: %v", event, data)
	eb.lock.RLock()
	handlers := append([]EventHandler(nil), eb.listeners[event]...)
	for _, ph := range eb.patternHandlers {
		if matchPattern(ph.pattern, event) {
			handlers = append(handlers, ph.handler)
		}
	}
	eb.lock.RUnlock()
	msg := &EventMessage{
		ID:        uuid.NewString(),
		Timestamp: time.Now(),
		Data:      data,
	}
	for _, handler := range handlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					// 可加日志：fmt.Printf("event handler panic: %v\n", r)
				}
			}()
			h(msg)
		}(handler)
	}
}

// Publish 支持 context、超时、日志。
func (eb *GenericEventBus[T]) Publish(ctx context.Context, event string, data T, timeout time.Duration) {
	// lib.Log.Debugf("Publish event: %s, data: %v", event, data)
	eb.lock.RLock()
	handlers := append([]EventHandlerWithContext[T](nil), eb.listeners[event]...)
	for _, ph := range eb.patternHandlers {
		if matchPattern(ph.pattern, event) {
			handlers = append(handlers, ph.handler)
		}
	}
	eb.lock.RUnlock()
	for _, handler := range handlers {
		go func(h EventHandlerWithContext[T]) {
			ctxToUse := ctx
			if timeout > 0 {
				var cancel context.CancelFunc
				ctxToUse, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}
			defer func() {
				if r := recover(); r != nil && eb.logger != nil {
					eb.logger(event, "panic", map[string]interface{}{"error": r})
				}
			}()
			h(ctxToUse, data)
		}(handler)
	}
	if eb.logger != nil {
		eb.logger(event, "publish", map[string]interface{}{"count": len(handlers)})
	}
}

// SyncPublish triggers all handlers synchronously (in the current goroutine).
// If any handler panic, it will be recovered.
func (eb *EventBus) SyncPublish(event string, data interface{}) {
	// lib.Log.Debugf("SyncPublish event: %s, data: %v", event, data)
	eb.lock.RLock()
	handlers := append([]EventHandler(nil), eb.listeners[event]...)
	for _, ph := range eb.patternHandlers {
		if matchPattern(ph.pattern, event) {
			handlers = append(handlers, ph.handler)
		}
	}
	eb.lock.RUnlock()
	msg := &EventMessage{
		ID:        uuid.NewString(),
		Timestamp: time.Now(),
		Data:      data,
	}
	for _, handler := range handlers {
		func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					// 可加日志：fmt.Printf("event handler panic: %v\n", r)
				}
			}()
			h(msg)
		}(handler)
	}
}

// matchPattern 支持简单的前缀通配符（如 "user.*" 匹配 "user.create"）
func matchPattern(pattern, event string) bool {
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(event) >= len(prefix) && event[:len(prefix)] == prefix
	}
	return pattern == event
}

// Close removes all event listeners.
func (eb *EventBus) Close() {
	eb.lock.Lock()
	defer eb.lock.Unlock()
	eb.listeners = make(map[string][]EventHandler)
	eb.patternHandlers = nil
}

// HasSubscribers returns true if the event has any subscribers.
func (eb *EventBus) HasSubscribers(event string) bool {
	eb.lock.RLock()
	defer eb.lock.RUnlock()
	handlers, ok := eb.listeners[event]
	return ok && len(handlers) > 0
}

// EventStats returns a map of event names to their subscriber counts.
func (eb *EventBus) EventStats() map[string]int {
	stats := make(map[string]int)
	eb.lock.RLock()
	defer eb.lock.RUnlock()
	for evt, handlers := range eb.listeners {
		stats[evt] = len(handlers)
	}
	return stats
}

// getFuncPointer returns the pointer of a function for comparison.
func getFuncPointer(fn EventHandler) uintptr {
	return reflect.ValueOf(fn).Pointer()
}
