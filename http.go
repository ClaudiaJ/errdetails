package errdetails

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// Handler mimics http.Handler, but with custom handling of returned errors.
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

// HandlerFunc type is an adapter to allow the use of ordinary functions as HTTP handlers.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

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

// ServeHTTP serves a JSON error response back to client if the Handler would return an error.
func (fn HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if verr := fn(w, r); verr != nil {
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
					err = msg
				}
			}
			// turn error details into protobuf details
			if msg, ok := verr.(proto.Message); ok {
				if any, err := ptypes.MarshalAny(msg); err == nil {
					p.Details = append(p.Details, any)
				}
			}
			// unwrap and move on the next
			if verr = errors.Unwrap(verr); verr == nil {
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")

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
}
