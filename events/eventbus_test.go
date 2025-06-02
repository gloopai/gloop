package events

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestEventBus_SubscribeAndPublish(t *testing.T) {
	eb := NewEventBus()
	var wg sync.WaitGroup
	called := false

	wg.Add(1)
	eb.Subscribe("test_event", func(msg *EventMessage) {
		defer wg.Done()
		if str, ok := msg.Data.(string); ok && str == "hello" {
			called = true
		}
	})

	eb.Publish("test_event", "hello")
	wg.Wait()

	if !called {
		t.Error("event handler was not called or received wrong data")
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	eb := NewEventBus()
	var wg sync.WaitGroup
	called := false
	handler := func(msg *EventMessage) {
		called = true
	}
	eb.Subscribe("event", handler)
	eb.Unsubscribe("event", handler)
	eb.Publish("event", nil)
	// Give goroutine a chance to run
	wg.Add(1)
	go func() { defer wg.Done() }()
	wg.Wait()
	if called {
		t.Error("handler should not be called after unsubscribe")
	}
}

func TestEventBus_Once(t *testing.T) {
	eb := NewEventBus()
	var wg sync.WaitGroup
	count := 0
	wg.Add(1)
	eb.Once("once_event", func(msg *EventMessage) {
		count++
		wg.Done()
	})
	eb.Publish("once_event", nil)
	wg.Wait()
	eb.Publish("once_event", nil)
	if count != 1 {
		t.Errorf("handler should be called only once, got %d", count)
	}
}

func TestEventBus_Close(t *testing.T) {
	eb := NewEventBus()
	called := false
	eb.Subscribe("close_event", func(msg *EventMessage) { called = true })
	eb.Close()
	eb.Publish("close_event", nil)
	if called {
		t.Error("handler should not be called after Close")
	}
}

func TestEventBus_HasSubscribers(t *testing.T) {
	eb := NewEventBus()
	if eb.HasSubscribers("evt") {
		t.Error("should not have subscribers")
	}
	eb.Subscribe("evt", func(msg *EventMessage) {})
	if !eb.HasSubscribers("evt") {
		t.Error("should have subscribers")
	}
}

func TestEventBus_EventStats(t *testing.T) {
	eb := NewEventBus()
	eb.Subscribe("a", func(msg *EventMessage) {})
	eb.Subscribe("a", func(msg *EventMessage) {})
	eb.Subscribe("b", func(msg *EventMessage) {})
	stats := eb.EventStats()
	if stats["a"] != 2 || stats["b"] != 1 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}

func TestEventBus_SubscribeParamCheck(t *testing.T) {
	eb := NewEventBus()
	eb.Subscribe("", func(msg *EventMessage) {})
	eb.Subscribe("evt", nil)
	if eb.HasSubscribers("") || eb.HasSubscribers("evt") {
		t.Error("should not subscribe with empty event or nil handler")
	}
}

func TestEventBus_SubscribeDedup(t *testing.T) {
	eb := NewEventBus()
	count := 0
	h := func(msg *EventMessage) { count++ }
	eb.Subscribe("evt", h)
	eb.Subscribe("evt", h)
	eb.SyncPublish("evt", nil)
	if count != 1 {
		t.Errorf("handler should only be called once, got %d", count)
	}
}

func TestEventBus_SyncPublish(t *testing.T) {
	eb := NewEventBus()
	var order []int
	eb.Subscribe("evt", func(msg *EventMessage) { order = append(order, 1) })
	eb.Subscribe("evt", func(msg *EventMessage) { order = append(order, 2) })
	eb.SyncPublish("evt", nil)
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Errorf("SyncPublish should call handlers in order, got %v", order)
	}
}

func TestEventBus_SubscribePattern(t *testing.T) {
	eb := NewEventBus()
	var mu sync.Mutex
	var called []string
	eb.SubscribePattern("user.*", func(msg *EventMessage) {
		mu.Lock()
		if str, ok := msg.Data.(string); ok {
			called = append(called, str)
		}
		mu.Unlock()
	})
	var wg sync.WaitGroup
	wg.Add(2)
	eb.Subscribe("user.create", func(msg *EventMessage) { wg.Done() })
	eb.Subscribe("user.delete", func(msg *EventMessage) { wg.Done() })
	eb.Publish("user.create", "a")
	eb.Publish("user.delete", "b")
	wg.Wait()
	eb.SyncPublish("user.update", "c")
	mu.Lock()
	cnt := len(called)
	mu.Unlock()
	if cnt != 3 {
		t.Errorf("pattern handler should be called for all user.* events, got %v", called)
	}
}

func TestEventBus_UnsubscribePattern(t *testing.T) {
	eb := NewEventBus()
	count := 0
	h := func(msg *EventMessage) { count++ }
	eb.SubscribePattern("order.*", h)
	eb.UnsubscribePattern("order.*", h)
	eb.SyncPublish("order.create", nil)
	if count != 0 {
		t.Error("pattern handler should not be called after unsubscribe")
	}
}

func TestGenericEventBus_PublishWithTimeoutAndLogger(t *testing.T) {
	var logCalls []string
	logger := func(event, action string, info map[string]interface{}) {
		logCalls = append(logCalls, event+":"+action)
	}
	eb := NewGenericEventBus[string](logger)
	ctx := context.Background()
	var called []string
	eb.Subscribe("evt", func(ctx context.Context, data string) {
		called = append(called, data)
	})
	eb.Publish(ctx, "evt", "hello", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	if len(called) != 1 || called[0] != "hello" {
		t.Errorf("handler not called or wrong data: %v", called)
	}
	if len(logCalls) == 0 || logCalls[0] != "evt:subscribe" || logCalls[1] != "evt:publish" {
		t.Errorf("logger not called as expected: %v", logCalls)
	}
}

func TestGenericEventBus_Timeout(t *testing.T) {
	eb := NewGenericEventBus[string](nil)
	ctx := context.Background()
	ch := make(chan struct{})
	eb.Subscribe("evt", func(ctx context.Context, data string) {
		<-ctx.Done()
		ch <- struct{}{}
	})
	eb.Publish(ctx, "evt", "hi", 10*time.Millisecond)
	select {
	case <-ch:
		// ok
	case <-time.After(100 * time.Millisecond):
		t.Error("handler did not respect timeout")
	}
}
