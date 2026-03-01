// 注册 gRPC 服务
a.proxy.Node.AddServiceProvider(&pb.Greeter_ServiceDesc, &server{})

go func() {
	time.Sleep(2 * time.Second)
	client, err := a.proxy.GetServiceClient()
	if err != nil {
		lib.Log.Error(err)
		return
	}
	greeterClient := pb.NewGreeterClient(client)
	resp, err := greeterClient.SayHello(context.Background(), &pb.HelloRequest{Name: "World"})
	if err != nil {
		lib.Log.Error(err)
		return
	}
	println("Response from gRPC server:", resp.GetMessage())
}()