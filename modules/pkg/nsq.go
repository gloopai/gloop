package pkg

import (
	"github.com/gloopai/gloop/lib"
	"github.com/nsqio/go-nsq"
)

type NsqClientOptions struct {
	Producer  string // 生产者服务器
	Subscribe string // 消息订阅服务器
}

type NsqClient struct {
	Base
	producer *nsq.Producer
	consumer *nsq.Consumer
	options  NsqClientOptions
}

func NewNsqClient(options NsqClientOptions) *NsqClient {
	return &NsqClient{
		options: options,
	}
}

func (n *NsqClient) Name() string {
	return "nsq-client"
}

func (n *NsqClient) Init() {
	// 初始化 Producer
	if n.options.Producer != "" {
		config := nsq.NewConfig()
		producer, err := nsq.NewProducer(n.options.Producer, config)
		if err != nil {
			lib.Log.Errorf("Failed to create NSQ producer: %v", err)
		} else {
			n.producer = producer
		}
	}
}

func (n *NsqClient) Start() error {
	if n.producer != nil {
		// 生产者不需要额外的启动步骤
		lib.Log.Infof("NSQ producer is ready to publish messages")
	}

	// if n.consumer != nil {
	// 	// 消息消费者已经在 Subscribe 方法中连接到 NSQ
	// 	lib.Log.Infof("NSQ consumer is ready to receive messages")
	// }
	return nil
}

func (n *NsqClient) Close() {
	if n.producer != nil {
		n.producer.Stop()
	}
	if n.consumer != nil {
		n.consumer.Stop()
	}
}

func (n *NsqClient) Destroy() {
	n.Close()
}

// Publish 发布消息到指定 topic
func (n *NsqClient) Publish(topic string, message []byte) error {
	if n.producer == nil {
		return nsq.ErrNotConnected
	}
	return n.producer.Publish(topic, message)
}

// Subscribe 订阅指定 topic 的消息
// handler: 处理消息的回调函数
func (n *NsqClient) Subscribe(topic, channel string, handler nsq.Handler) error {
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		lib.Log.Errorf("Failed to create NSQ consumer: %v", err)
		return err
	}
	consumer.AddHandler(handler)
	if n.options.Subscribe == "" {
		return nsq.ErrNotConnected
	}
	err = consumer.ConnectToNSQD(n.options.Subscribe)
	if err != nil {
		lib.Log.Errorf("Failed to connect to NSQD: %v", err)
		return err
	}
	n.consumer = consumer
	return nil
}
