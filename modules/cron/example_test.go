package cron_test

import (
	"fmt"
	"time"

	"github.com/gloopai/gloop/modules/cron"
)

// 此示例展示 cron 封装的基本用法。
func Example_basic() {
	c := cron.NewCron(cron.CronOptions{UseSeconds: true, Location: "UTC"})

	id, _ := c.AddFunc("@every 2s", "tick", func() {
		fmt.Println("tick")
	})

	c.Start()

	time.Sleep(3 * time.Second)

	c.Remove(id)
	ctx := c.Stop()
	<-ctx.Done()

	// Output:
	// tick
}
