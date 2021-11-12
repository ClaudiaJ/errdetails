package errdetails

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

const contentType = "application/json"

// HandlerFunc type is an adapter to allow the use of ordinary functions as HTTP handlers.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// ServeHTTP serves a JSON error response back to client if the Handler would return an error.
//
// Note of caution: Masking or otherwise distinguishing details safe to share
// to end client is an exercise left to the implementor.
func (fn HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	verr := fn(w, r)
	if verr == nil {
		return
	}

	var s *status.Status

	// become a Status one way or another
	var sterr statusError
	if !errors.As(verr, &sterr) {
		sterr = &errCodeError{error: verr, Code: codes.Unknown}
	}

	p := status.Convert(sterr).Proto()
	for {
		// render localizable messages
		if loc, ok := verr.(localizable); ok {
			if msg, err := loc.Localize(r); err == nil {
				verr = msg
			}
		}
		// turn error details into protobuf details
		if msg, ok := verr.(protoreflect.ProtoMessage); ok {
			if any, err := anypb.New(msg); err == nil {
				p.Details = append(p.Details, any)
			}
		}
		// unwrap and move on the next
		if verr = errors.Unwrap(verr); verr == nil {
			break
		}
	}

	w.Header().Set("Content-Type", contentType)

	var resp json.RawMessage
	enc := json.NewEncoder(w)

	resp, err := protojson.Marshal(p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = enc.Encode(map[string]interface{}{
			"status":  codes.Internal,
			"message": "Internal Server Error: failed to encode error response",
		})
		return
	}

	w.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))
	_ = enc.Encode(resp)
}

// DetailsMapper provides a mapping from arbitrary Protobuf message type to an
// Error wrapper that will reconstruct a fully wrapped error type from JSON data.
//
// This enables error types built ontop protobuf messages not provided within
// this package to be reconstructed from http response body the same as all of
// the error types provided by this module.
//
// This will go away whenever I can figure out how to acheive this with protoreflect.
type DetailsMapper interface {
	Map(protoreflect.ProtoMessage) Details
}

// FromJSON reads JSON fom a Reader like a response Body, and makes best effort
// to reconstruct the wrapped error from gRPC Status such that errors.As and
// errors.Is may still be satisfied by the error interface types.
//
// For any expected error not already accomodated by this package, you can
// provide optional DetailsMappers.
//
// If the Map method of a DetailsMapper returns an implementation of Details
// wrapper, the error is further wrapped by the mapped wrapper.
func FromJSON(r io.Reader, mappers ...DetailsMapper) error {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}

	s := &statuspb.Status{}
	if err := protojson.Unmarshal(buf.Bytes(), s); err != nil {
		return err
	}

	sterr := New(codes.Code(s.Code), s.Message)

	for _, detail := range s.Details {
		pb, err := anypb.UnmarshalNew(detail, proto.UnmarshalOptions{})
		if err != nil {
			return err
		}
		// consider arbitrary client-provided error types too
		// TODO: How to better leverage protoreflect?
		for _, mapper := range mappers {
			if wrapper := mapper.Map(pb); wrapper != nil {
				sterr = wrapper.Wrap(sterr)
			}
		}

		switch msg := pb.(type) {
		case *errdetails.BadRequest:
			sterr = &errBadRequest{error: sterr, BadRequest: msg}
		case *errdetails.DebugInfo:
			sterr = &errDebugInfo{error: sterr, DebugInfo: msg}
		case *errdetails.ErrorInfo:
			sterr = &errInfo{error: sterr, ErrorInfo: msg}
		case *errdetails.Help:
			sterr = &errHelpLink{error: sterr, Help: msg}
		case *errdetails.LocalizedMessage:
			sterr = &localizedError{error: sterr, LocalizedMessage: msg}
		case *errdetails.PreconditionFailure:
			sterr = &errPreconditionFailed{error: sterr, PreconditionFailure: msg}
		case *errdetails.QuotaFailure:
			sterr = &errQuotaFailure{error: sterr, QuotaFailure: msg}
		case *errdetails.RequestInfo:
			sterr = &errRequestInfo{error: sterr, RequestInfo: msg}
		case *errdetails.ResourceInfo:
			sterr = &errResourceInfo{error: sterr, ResourceInfo: msg}
		case *errdetails.RetryInfo:
			sterr = &errRetryInfo{error: sterr, RetryInfo: msg}
		default:
			sterr = WithDetails(sterr, wrapperFunc(func(err error) error {
				return &arbitraryError{error: err, ProtoMessage: msg}
			}))
		}
	}

	return sterr
}

// arbitraryError just enables us to put our protobuf details back into some
// error type without losing it.
//
// arbitrary only embeds a ProtoMessage, and it's up to the end user to know
// what the  heck to do with that.
type arbitraryError struct {
	error
	protoreflect.ProtoMessage
}

// Is implements errors.Is interface.
func (e *arbitraryError) Is(target error) bool {
	if v, ok := target.(protoreflect.ProtoMessage); ok {
		// tbh idk what I'm doing here yet
		return v.ProtoReflect().Interface() == e.ProtoReflect().Interface()
	}

	return false
}

// Unwrap implements Unwrap.
func (e *arbitraryError) Unwrap() error {
	return e.error
}

// statusError is a neat little trick used in gRPC status module to enable
// errors to self-describe a conversion to Status.
type statusError interface {
	error
	GRPCStatus() *status.Status
}

type localizable interface {
	// Localize renders a localizable error message to the client-requested locale.
	Localize(r *http.Request) (LocalizedError, error)
}
