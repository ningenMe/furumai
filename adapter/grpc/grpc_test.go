package grpc

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// startEchoServer starts a minimal in-process gRPC server exposing
// /test.Echo/Say, hand-registered via grpc.ServiceDesc so this test needs
// no .proto/codegen step. It echoes the request string back, prefixed.
func startEchoServer(t *testing.T) (addr string, stop func()) {
	t.Helper()

	desc := grpc.ServiceDesc{
		ServiceName: "test.Echo",
		HandlerType: (*any)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Say",
				Handler: func(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
					req := new(wrapperspb.StringValue)
					if err := dec(req); err != nil {
						return nil, err
					}
					if req.Value == "fail" {
						return nil, status.Error(codes.InvalidArgument, "bad input")
					}
					grpc.SetTrailer(ctx, metadata.Pairs("x-test", "1"))
					return &wrapperspb.StringValue{Value: "echo: " + req.Value}, nil
				},
			},
		},
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	srv.RegisterService(&desc, nil)
	go srv.Serve(lis)

	return lis.Addr().String(), srv.Stop
}

func TestUnarySuccess(t *testing.T) {
	addr, stop := startEchoServer(t)
	defer stop()

	client, err := NewStimulus(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("NewStimulus: %v", err)
	}
	defer client.Close()

	reply := &wrapperspb.StringValue{}
	resp, err := client.Unary(context.Background(), "/test.Echo/Say", &wrapperspb.StringValue{Value: "hi"}, reply)
	if err != nil {
		t.Fatalf("Unary: %v", err)
	}

	if resp.StatusCode != codes.OK {
		t.Errorf("StatusCode = %v, want %v", resp.StatusCode, codes.OK)
	}
	if got := reply.GetValue(); got != "echo: hi" {
		t.Errorf("reply.Value = %q, want %q", got, "echo: hi")
	}
}

func TestUnaryErrorStatus(t *testing.T) {
	addr, stop := startEchoServer(t)
	defer stop()

	client, err := NewStimulus(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("NewStimulus: %v", err)
	}
	defer client.Close()

	reply := &wrapperspb.StringValue{}
	resp, err := client.Unary(context.Background(), "/test.Echo/Say", &wrapperspb.StringValue{Value: "fail"}, reply)
	if err != nil {
		t.Fatalf("Unary: %v", err)
	}

	if resp.StatusCode != codes.InvalidArgument {
		t.Errorf("StatusCode = %v, want %v", resp.StatusCode, codes.InvalidArgument)
	}
}
