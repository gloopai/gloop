简单的定时器封装，基于 github.com/robfig/cron/v3

快速使用：

```go
import (
    "time"
    "github.com/gloopai/gloop/modules/cron"
)

func main() {
    c := cron.NewCron(cron.CronOptions{UseSeconds:true, Location: "Asia/Shanghai"})
    c.AddFunc("@every 1m", "job-name", func(){
        // do work
    })
    c.Start()
    defer func(){ <-c.Stop().Done() }()
    time.Sleep(2 * time.Minute)
}
```

说明：
- `NewCron` 支持秒级表达式（UseSeconds）和时区（Location）。
- 提供 `AddFunc`、`AddJob`、`Remove`、`Entries`、`Start`、`Stop`。

附加示例（一次性与立即执行）：

ScheduleAt 在指定时间点执行一次：

```go
func exampleScheduleAt() {
    c := cron.NewCron(cron.CronOptions{UseSeconds:true, Location: "UTC"})
    c.Start()
    // 在 10 秒后执行一次
    t := time.Now().Add(10 * time.Second)
    id, _ := c.ScheduleAt(t, "single-job", func(){
        fmt.Println("ran at", time.Now())
    })
    // 可选：如果想取消
    // c.Remove(id)
    <-time.After(12 * time.Second)
    <-c.Stop().Done()
}
```

RunOnce 立即在后台执行一次任务：

```go
func exampleRunOnce() {
    c := cron.NewCron(cron.CronOptions{UseSeconds:true})
    // 不需要 Start/Stop 也能运行一次性任务，但推荐 Start/Stop 管理生命周期
    c.Start()
    c.RunOnce("now-job", func(){
        fmt.Println("ran once immediately")
    })
    <-time.After(1 * time.Second)
    <-c.Stop().Done()
}
```

安装依赖：

在仓库根目录运行：

```sh
go get github.com/robfig/cron/v3
go mod tidy
```
