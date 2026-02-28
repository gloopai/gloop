package pkg

type NsqClientOptions struct {
}

type NsqClient struct {
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

func (n *NsqClient) SetEnv(env *interface{}) {}

func (n *NsqClient) GetEnv() *interface{} {
	return nil
}
