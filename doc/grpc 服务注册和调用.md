env := a.GetEnv()
env.Node.AddServiceProvider("Test", &pb.Greeter_ServiceDesc, &server{})

go func() {
    time.Sleep(2 * time.Second)
    client, err := env.Node.ServiceClient("Test")
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