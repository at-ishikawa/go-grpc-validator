package grpc_validator

import (
    `context`
    `fmt`
    `log`
    `net`

    `google.golang.org/genproto/googleapis/rpc/errdetails`
    `google.golang.org/grpc`
    `google.golang.org/grpc/status`
    echo `github.com/at-ishikawa/go-grpc-validator/testdata/proto`
    grpc_playground_validator `github.com/at-ishikawa/go-grpc-validator/playground/v9`
)

// server is used to implement helloworld.GreeterServer.
type server struct {
    echo.UnimplementedEchoServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *echo.EchoRequest) (*echo.EchoResponse, error) {
	log.Printf("Received: %v", in.GetMessage())
	return &echo.EchoResponse{Message: "Hello " + in.GetMessage()}, nil
}

func ExamplePlaygroundValidatorUnaryServerInterceptor() {
    v, err := grpc_playground_validator.NewValidator()
    if err != nil {
        log.Fatal(err)
    }
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(
	    grpc_playground_validator.UnaryServerInterceptor(v),
    ))
	go func() {
        echo.RegisterEchoServer(s, &server{})
        if err := s.Serve(lis); err != nil {
            log.Fatalf("failed to serve: %v", err)
        }
    }()

	ctx := context.Background()
    dial, err := grpc.DialContext(ctx, ":50051", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer func() {
        if err := dial.Close(); err != nil {
            log.Println(err)
        }
    }()
    client := echo.NewEchoClient(dial)
    res, err := client.UnaryEcho(ctx, &echo.EchoRequest{
        Message: "",
    })
    if err == nil {
        log.Fatalf("want error but got %v", res)
    }
    st, ok := status.FromError(err)
    if !ok {
        log.Fatalf("want status, but got %v", err)
    }
    fmt.Printf("Code:: %v\n", st.Code())
    fmt.Printf("Message:: %s\n", st.Message())
    fmt.Printf("Details:: ")
    for _, d := range st.Details() {
        br, ok := d.(*errdetails.BadRequest)
        if !ok {
            log.Fatal(d)
        }
        for _, fv := range br.FieldViolations {
            fmt.Printf("%s: %s", fv.Field, fv.Description)
        }
    }
	s.GracefulStop()

    // Output:
    // Code:: InvalidArgument
    // Message:: failed to validate request
    // Details:: EchoRequest.Message: Message must be at least 8 characters in length
}
