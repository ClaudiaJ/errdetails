package errdetails

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
)

func TestHandler(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)

	handler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return New(codes.InvalidArgument, "test error",
			Help(&errdetails.Help_Link{
				Url:         "url2",
				Description: "desc4",
			}, &errdetails.Help_Link{
				Url:         "url1",
				Description: "desc3",
			}),
			BadRequest(&errdetails.BadRequest_FieldViolation{
				Field:       "field1",
				Description: "desc1",
			}, &errdetails.BadRequest_FieldViolation{
				Field:       "field2",
				Description: "desc2",
			}),
		)
	})
	handler.ServeHTTP(rr, req)

	// code 3 == ErrInvalidArgument
	// firt status Code is applied in order, further status codes are informational
	exp := `{
		"code": 3,
		"message": "InvalidArgument: test error",
		"details": [{
			"@type": "type.googleapis.com/google.rpc.BadRequest",
			"fieldViolations": [{
				"description": "desc1",
				"field": "field1"
			}, {
				"description": "desc2",
				"field": "field2"
			}]
		}, {
			"@type": "type.googleapis.com/google.rpc.Help",
			"links": [{
				"description": "desc4",
				"url": "url2"
			}, {
				"description": "desc3",
				"url": "url1"
			}]
		}]
	}`
	require.JSONEq(t, exp, rr.Body.String())
}
