package pkg

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gloopai/gloop/lib"
	"github.com/nsqio/go-nsq"
)

var (
	ErrProducerNotInitialized = errors.New("producer not initialized")
	ErrInvalidConfig          = errors.New("invalid configuration")
)

type NsqClientOptions struct {
	Producer          string        // 生产者服务器地址
	Subscribe         string        // 消息订阅服务器地址
	MaxInFlight       int           // 最大并发处理消息数
	ConnectTimeout    time.Duration // 连接超时时间
	ReadTimeout       time.Duration // 读取超时时间
	WriteTimeout      time.Duration // 写入超时时间
	AutoReconnect     bool          // 是否自动重连
	MaxReconnectDelay time.Duration // 最大重连延迟
}

type NsqClient struct {
	Base
	producer  *nsq.Producer
	consumers map[string]*nsq.Consumer // key: topic:channel
	options   NsqClientOptions
	mu        sync.RWMutex // 保护consumers map的并发访问
	ctx       context.Context
	cancel    context.CancelFunc
	isStarted bool
}

func NewNsqClient(options NsqClientOptions) *NsqClient {
	// 设置默认值
	if options.MaxInFlight <= 0 {
		options.MaxInFlight = 10
	}
	if options.ConnectTimeout <= 0 {
		options.ConnectTimeout = 10 * time.Second
	}
	if options.ReadTimeout <= 0 {
		options.ReadTimeout = 60 * time.Second
	}
	if options.WriteTimeout <= 0 {
		options.WriteTimeout = 5 * time.Second
	}
	if options.MaxReconnectDelay <= 0 {
		options.MaxReconnectDelay = 30 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &NsqClient{
		options:   options,
		consumers: make(map[string]*nsq.Consumer),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (n *NsqClient) Name() string {
	return "nsq-client"
}

// createNSQConfig 创建NSQ配置
func (n *NsqClient) createNSQConfig() *nsq.Config {
	config := nsq.NewConfig()
	config.MaxInFlight = n.options.MaxInFlight
	config.DialTimeout = n.options.ConnectTimeout
	config.ReadTimeout = n.options.ReadTimeout
	config.WriteTimeout = n.options.WriteTimeout
	// NSQ config doesn't have MaxReconnectDelay field, using MaxBackoffDuration instead
	config.MaxBackoffDuration = n.options.MaxReconnectDelay
	config.DefaultRequeueDelay = 0
	config.BackoffStrategy = &nsq.ExponentialStrategy{}

	return config
}

func (n *NsqClient) Init() error {
	// 验证配置
	if n.options.Producer == "" && n.options.Subscribe == "" {
		return fmt.Errorf("%w: both producer and subscribe addresses cannot be empty", ErrInvalidConfig)
	}

	// 初始化 Producer
	if n.options.Producer != "" {
		config := n.createNSQConfig()
		producer, err := nsq.NewProducer(n.options.Producer, config)
		if err != nil {
			return fmt.Errorf("failed to create NSQ producer: %w", err)
		}
		n.producer = producer
		lib.Log.Infof("NSQ producer initialized for address: %s", n.options.Producer)
	}

	return nil
}

func (n *NsqClient) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.isStarted {
		return nil
	}

	if n.producer != nil {
		// 测试生产者连接
		err := n.producer.Ping()
		if err != nil {
			return fmt.Errorf("failed to ping NSQ producer: %w", err)
		}
		lib.Log.Infof("NSQ producer is ready to publish messages")
	}

	n.isStarted = true
	return nil
}

func (n *NsqClient) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isStarted {
		return nil
	}

	// 停止生产者
	if n.producer != nil {
		n.producer.Stop()
		lib.Log.Infof("NSQ producer stopped")
	}

	// 停止所有消费者
	for key, consumer := range n.consumers {
		consumer.Stop()
		<-consumer.StopChan // 等待消费者完全停止
		lib.Log.Infof("NSQ consumer stopped: %s", key)
	}

	// 取消上下文
	if n.cancel != nil {
		n.cancel()
	}

	n.isStarted = false
	return nil
}

func (n *NsqClient) Destroy() error {
	return n.Close()
}

// Publish 发布消息到指定 topic
func (n *NsqClient) Publish(topic string, message []byte) error {
	if topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}
	if len(message) == 0 {
		return fmt.Errorf("message cannot be empty")
	}

	n.mu.RLock()
	producer := n.producer
	started := n.isStarted
	n.mu.RUnlock()

	if producer == nil {
		return ErrProducerNotInitialized
	}
	if !started {
		return fmt.Errorf("nsq client not started")
	}

	return producer.Publish(topic, message)
}

// PublishAsync 异步发布消息
func (n *NsqClient) PublishAsync(topic string, message []byte, doneChan chan *nsq.ProducerTransaction, args ...interface{}) error {
	if topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}
	if len(message) == 0 {
		return fmt.Errorf("message cannot be empty")
	}

	n.mu.RLock()
	producer := n.producer
	started := n.isStarted
	n.mu.RUnlock()

	if producer == nil {
		return ErrProducerNotInitialized
	}
	if !started {
		return fmt.Errorf("nsq client not started")
	}

	return producer.PublishAsync(topic, message, doneChan, args...)
}

// Subscribe 订阅指定 topic 的消息
// handler: 处理消息的回调函数
func (n *NsqClient) Subscribe(topic, channel string, handler nsq.Handler) error {
	if topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}
	if channel == "" {
		return fmt.Errorf("channel cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	if n.options.Subscribe == "" {
		return fmt.Errorf("subscribe address not configured")
	}

	key := topic + ":" + channel

	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.consumers[key]; exists {
		lib.Log.Warnf("Already subscribed to topic=%s channel=%s", topic, channel)
		return nil
	}

	config := n.createNSQConfig()
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return fmt.Errorf("failed to create NSQ consumer for topic=%s channel=%s: %w", topic, channel, err)
	}

	consumer.AddHandler(handler)

	err = consumer.ConnectToNSQD(n.options.Subscribe)
	if err != nil {
		consumer.Stop() // 清理资源
		return fmt.Errorf("failed to connect to NSQD %s for topic=%s channel=%s: %w", n.options.Subscribe, topic, channel, err)
	}

	n.consumers[key] = consumer
	lib.Log.Infof("Successfully subscribed to topic=%s channel=%s", topic, channel)
	return nil
}

// Unsubscribe 取消订阅
func (n *NsqClient) Unsubscribe(topic, channel string) error {
	if topic == "" || channel == "" {
		return fmt.Errorf("topic and channel cannot be empty")
	}

	key := topic + ":" + channel

	n.mu.Lock()
	defer n.mu.Unlock()

	consumer, exists := n.consumers[key]
	if !exists {
		return fmt.Errorf("not subscribed to topic=%s channel=%s", topic, channel)
	}

	consumer.Stop()
	<-consumer.StopChan // 等待消费者完全停止
	delete(n.consumers, key)

	lib.Log.Infof("Unsubscribed from topic=%s channel=%s", topic, channel)
	return nil
}

// IsConnected 检查连接状态
func (n *NsqClient) IsConnected() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.producer != nil {
		return n.producer.Ping() == nil
	}

	// 如果只有消费者，检查至少一个消费者是否连接
	for _, consumer := range n.consumers {
		stats := consumer.Stats()
		if stats != nil && stats.Connections > 0 {
			return true
		}
	}

	return false
}

// GetStats 获取统计信息
func (n *NsqClient) GetStats() map[string]interface{} {
	n.mu.RLock()
	defer n.mu.Unlock()

	stats := make(map[string]interface{})
	stats["started"] = n.isStarted
	stats["producer_connected"] = n.producer != nil && n.producer.Ping() == nil
	stats["consumers_count"] = len(n.consumers)

	consumerStats := make(map[string]interface{})
	for key, consumer := range n.consumers {
		consumerStats[key] = consumer.Stats()
	}
	stats["consumers"] = consumerStats

	return stats
}
