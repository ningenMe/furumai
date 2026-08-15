// Package grpc is furumai's gRPC Stimulus/Observation adapter.
//
// Note: this package is conventionally imported as "grpc", the same as
// google.golang.org/grpc itself. Callers that need both in one file (to
// pass grpc.DialOption/grpc.CallOption values, as most will) should alias
// one of the two imports, e.g. `googlegrpc "google.golang.org/grpc"`.
package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Stimulus invokes gRPC methods against Conn. It is used from both given
// and when steps, since Stimulus adapters are shared between them.
type Stimulus struct {
	Conn *grpc.ClientConn
}

// NewStimulus dials target (see google.golang.org/grpc.NewClient) and
// returns a Stimulus. Credentials must be supplied via opts (e.g.
// grpc.WithTransportCredentials(insecure.NewCredentials()) for plaintext).
func NewStimulus(target string, opts ...grpc.DialOption) (*Stimulus, error) {
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, err
	}
	return &Stimulus{Conn: conn}, nil
}

// Close closes the underlying connection.
func (s *Stimulus) Close() error { return s.Conn.Close() }

// Response is the full-state Observation for a unary RPC call. Message is
// the caller's own proto.Message type; StatusCode is a
// google.golang.org/grpc/codes.Code. All three are typed any so a
// furumai.Matcher (Any, Regex, ...) can be substituted for any of them.
type Response struct {
	Message    any
	StatusCode any
	Trailer    any
}

// Unary invokes a unary RPC. method is the fully qualified method name
// (e.g. "/greeter.Greeter/SayHello"); req and reply are the caller's own
// generated proto.Message types, with reply populated on return. A non-OK
// status is reported via Response.StatusCode, not a Go error: err is only
// non-nil when the call couldn't be evaluated as an RPC outcome at all
// (e.g. it never reached the server).
func (s *Stimulus) Unary(ctx context.Context, method string, req, reply proto.Message, opts ...grpc.CallOption) (*Response, error) {
	var trailer metadata.MD
	err := s.Conn.Invoke(ctx, method, req, reply, append(opts, grpc.Trailer(&trailer))...)

	st, ok := status.FromError(err)
	if !ok {
		return nil, err
	}

	return &Response{
		Message:    reply,
		StatusCode: st.Code(),
		Trailer:    trailer,
	}, nil
}
