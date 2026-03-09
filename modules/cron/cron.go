package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type CronOptions struct {
	// Location 指定 IANA 时区名称，例如 "Asia/Shanghai"。为空则使用 UTC
	Location string
	// UseSeconds 启用秒级别的 cron 表达式。如果为 false，则秒字段为可选。
	UseSeconds bool
}

type EntryID = cron.EntryID

type EntryInfo struct {
	ID       EntryID
	Name     string
	Next     time.Time
	Prev     time.Time
	Schedule string
}

type Cron struct {
	mu           sync.RWMutex
	c            *cron.Cron
	entries      map[EntryID]string
	opts         CronOptions
	nextCustomID EntryID
	oneOffCancel map[EntryID]context.CancelFunc
	oneOffNext   map[EntryID]time.Time
}

func NewCron(options CronOptions) *Cron {
	var parser cron.Parser
	if options.UseSeconds {
		parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	} else {
		parser = cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	}

	loc := time.UTC
	if options.Location != "" {
		if l, err := time.LoadLocation(options.Location); err == nil {
			loc = l
		}
	}

	c := cron.New(cron.WithParser(parser), cron.WithLocation(loc))
	return &Cron{
		c:            c,
		entries:      make(map[EntryID]string),
		opts:         options,
		nextCustomID: EntryID(-1),
		oneOffCancel: make(map[EntryID]context.CancelFunc),
		oneOffNext:   make(map[EntryID]time.Time),
	}
}

// Start 在独立的 goroutine 中启动调度器。
func (cr *Cron) Start() {
	cr.c.Start()
}

// Stop 停止调度器，并返回一个在正在运行的任务完成后关闭的 context。
func (cr *Cron) Stop() context.Context {
	return cr.c.Stop()
}

// AddFunc 添加一个带可选名称的定时函数。名称可以为空。
func (cr *Cron) AddFunc(spec string, name string, cmd func()) (EntryID, error) {
	id, err := cr.c.AddFunc(spec, func() {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		cmd()
	})
	if err != nil {
		return 0, err
	}
	cr.mu.Lock()
	cr.entries[id] = name
	cr.mu.Unlock()
	return id, nil
}

// EverySeconds 每隔指定时长执行一次（支持秒精度），等价于 AddFunc("@every <dur>")。
func (cr *Cron) EverySeconds(d time.Duration, name string, cmd func()) (EntryID, error) {
	return cr.AddFunc("@every "+d.String(), name, cmd)
}

// EveryMinutes 每隔指定分钟执行一次。
func (cr *Cron) EveryMinutes(minutes int, name string, cmd func()) (EntryID, error) {
	return cr.EverySeconds(time.Duration(minutes)*time.Minute, name, cmd)
}

func (cr *Cron) allocCustomID() EntryID {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	id := cr.nextCustomID
	cr.nextCustomID--
	return id
}

// ScheduleAt 在指定时间点执行一次任务（单次调度）。如果时间已到或在过去，将立即执行一次。
func (cr *Cron) ScheduleAt(t time.Time, name string, cmd func()) (EntryID, error) {
	ctx, cancel := context.WithCancel(context.Background())
	eid := cr.allocCustomID()

	cr.mu.Lock()
	cr.entries[eid] = name
	cr.oneOffCancel[eid] = cancel
	cr.oneOffNext[eid] = t
	cr.mu.Unlock()

	go func() {
		var timer *time.Timer
		now := time.Now()
		if !t.After(now) {
			timer = time.NewTimer(0)
		} else {
			timer = time.NewTimer(time.Until(t))
		}
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			defer func() {
				if r := recover(); r != nil {
				}
			}()
			cmd()
			cr.mu.Lock()
			delete(cr.oneOffCancel, eid)
			delete(cr.oneOffNext, eid)
			delete(cr.entries, eid)
			cr.mu.Unlock()
		}
	}()
	return eid, nil
}

// RunOnce 立即在后台执行一次任务，返回可用于取消的 EntryID（如果需要）。
func (cr *Cron) RunOnce(name string, cmd func()) EntryID {
	id, _ := cr.ScheduleAt(time.Now(), name, cmd)
	return id
}

// AddJob 添加一个实现了 cron.Job 的任务对象。
func (cr *Cron) AddJob(spec string, name string, job cron.Job) (EntryID, error) {
	id, err := cr.c.AddJob(spec, job)
	if err != nil {
		return 0, err
	}
	cr.mu.Lock()
	cr.entries[id] = name
	cr.mu.Unlock()
	return id, nil
}

// Remove 根据 id 移除已调度的任务。
func (cr *Cron) Remove(id EntryID) {
	cr.mu.Lock()
	// 如果是一次性任务，调用取消并清理
	if cancel, ok := cr.oneOffCancel[id]; ok {
		cancel()
		delete(cr.oneOffCancel, id)
		delete(cr.oneOffNext, id)
		delete(cr.entries, id)
		cr.mu.Unlock()
		return
	}
	cr.mu.Unlock()
	// 否则当作 cron.EntryID 处理
	cr.c.Remove(id)
	cr.mu.Lock()
	delete(cr.entries, id)
	cr.mu.Unlock()
}

// Entries 返回当前已调度任务的基本信息列表。
func (cr *Cron) Entries() []EntryInfo {
	cronEntries := cr.c.Entries()
	out := make([]EntryInfo, 0, len(cronEntries))
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	for _, e := range cronEntries {
		name := cr.entries[e.ID]
		sched := ""
		if e.Schedule != nil {
			sched = fmt.Sprintf("%T", e.Schedule)
		}
		out = append(out, EntryInfo{
			ID:       e.ID,
			Name:     name,
			Next:     e.Next,
			Prev:     e.Prev,
			Schedule: sched,
		})
	}
	// 添加一次性任务的信息
	for id, t := range cr.oneOffNext {
		name := cr.entries[id]
		out = append(out, EntryInfo{
			ID:       id,
			Name:     name,
			Next:     t,
			Prev:     time.Time{},
			Schedule: "one-off",
		})
	}
	return out
}
