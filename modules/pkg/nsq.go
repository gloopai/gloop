package pkg

type NsqClientOptions struct {
	Producer  string // 生产者服务器
	Subscribe string // 消息订阅服务器
}

type NsqClient struct {
	Base
}

func NewNsqClient(options NsqClientOptions) *NsqClient {
	return &NsqClient{}
}

func (n *NsqClient) Name() string {
	return "nsq-client"
}

func (n *NsqClient) Init() {}

func (n *NsqClient) Start() error {
	return nil
}

func (n *NsqClient) Close() {}

func (n *NsqClient) Destroy() {}
