package modules

import (
	"context"
	"fmt"
	"sync"
	"time"

	gredis "github.com/redis/go-redis/v9"
)

type RdbOptions struct {
	Addr        string // 服务器 0.0.0.0:6379
	Password    string // 密码
	Db          int    // 数据库 数字 默认 0
	Prefix      string // key 前缀
	MaxIdle     int    //最初的连接数量 默认 20
	MaxActive   int    //连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配，默认0
	IdleTimeout int    // 连接关闭时间 more300
}

type Rdb struct {
	Base
	Config      *RdbOptions
	redisClient *gredis.Client
	poolMu      sync.Mutex
}

func NewRdb(opts *RdbOptions) *Rdb {
	return &Rdb{
		Config: opts,
	}
}

func (r *Rdb) createRedisConnect() {
	// 防止并发初始化
	r.poolMu.Lock()
	if r.redisClient != nil {
		r.poolMu.Unlock()
		return
	}
	defer r.poolMu.Unlock()

	maxIdleConns := 20
	if r.Config.MaxIdle != 0 {
		maxIdleConns = r.Config.MaxIdle
	}
	maxActive := 0
	if r.Config.MaxActive != 0 {
		maxActive = r.Config.MaxActive
	}
	opts := &gredis.Options{
		Addr:         r.Config.Addr,
		Password:     r.Config.Password,
		DB:           r.Config.Db,
		MinIdleConns: maxIdleConns,
		PoolSize:     maxActive,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	r.redisClient = gredis.NewClient(opts)
}

func (r *Rdb) Name() string {
	return "rdb"
}

func (r *Rdb) Init() {
	// r.printInfo()
}

func (r *Rdb) Start() error {
	if r.Config == nil {
		return fmt.Errorf("rdb config is nil")
	}
	// 初始化连接池（幂等）
	r.createRedisConnect()
	if r.redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := r.redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %v", err)
	}
	return nil
}

func (r *Rdb) Close() {
	if r.redisClient != nil {
		_ = r.redisClient.Close()
		r.redisClient = nil
	}
}

func (r *Rdb) GetConn() *gredis.Client {
	return r.redisClient
}
