package errdetails

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

// assert UnaryServerInterceptor is of the same type UnaryServerInterceptor
var _ grpc.UnaryServerInterceptor = UnaryServerInterceptor

// UnaryServerInterceptor transcribes wrapped errors with details into gRPC Status.
func UnaryServerInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)

	return resp, translateError(err)
}

// assert StreamServerInterceptor is of the same type StreamServerInterceptor
var _ grpc.StreamServerInterceptor = StreamServerInterceptor

// StreamServerInterceptor transcribes wrapped errors with details into gRPC Status.
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	return translateError(handler(srv, ss))
}

func translateError(err error) error {
	// become a Status one way or another
	var sterr statusError
	if !errors.As(err, &sterr) {
		sterr = &errCodeError{error: err, Code: codes.Unknown}
	}

	p := status.Convert(sterr).Proto()
	for {
		// turn error details into protobuf details
		if msg, ok := err.(protoreflect.ProtoMessage); ok {
			if any, err := anypb.New(msg); err == nil {
				p.Details = append(p.Details, any)
			}
		}
		// unwrap and move on the next
		if err = errors.Unwrap(err); err == nil {
			break
		}
	}

	return status.FromProto(p).Err()
}
