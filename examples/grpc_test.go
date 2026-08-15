package examples

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/ningenMe/furumai"
	furumaigrpc "github.com/ningenMe/furumai/adapter/grpc"
)

// TestEchoRPC demonstrates the gRPC Stimulus/Observation adapter: when
// invokes a unary RPC, then observes the full Response (message, status
// code) and compares it structurally against an expected value.
//
// The server here is hand-registered (no .proto/codegen) purely to keep
// the example self-contained; a real caller would use their own generated
// client/server stubs.
func TestEchoRPC(t *testing.T) {
	desc := grpc.ServiceDesc{
		ServiceName: "greeter.Greeter",
		HandlerType: (*any)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "SayHello",
				Handler: func(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
					req := new(wrapperspb.StringValue)
					if err := dec(req); err != nil {
						return nil, err
					}
					return &wrapperspb.StringValue{Value: "hello, " + req.Value}, nil
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
	defer srv.Stop()

	client, err := furumaigrpc.NewStimulus(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("NewStimulus: %v", err)
	}
	defer client.Close()

	reply := &wrapperspb.StringValue{}
	var resp *furumaigrpc.Response

	furumai.When(t, func() error {
		var err error
		resp, err = client.Unary(context.Background(), "/greeter.Greeter/SayHello", &wrapperspb.StringValue{Value: "Alice"}, reply)
		return err
	})

	furumai.ThenEqual(t, *resp, furumaigrpc.Response{
		Message:    &wrapperspb.StringValue{Value: "hello, Alice"},
		StatusCode: codes.OK,
		Trailer:    furumai.Ignore(),
	})
}
