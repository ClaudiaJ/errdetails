package errdetails

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
)

func TestHandler(t *testing.T) {
	testHandler(t)

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

func TestFromJSON(t *testing.T) {
	testHandler(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)

	handler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return New(codes.ResourceExhausted, "resend email rate limit exceeded",
			RetryDelay(time.Minute),
			QuotaFailure(&errdetails.QuotaFailure_Violation{
				Subject:     "auth0|123456789",
				Description: "Rate limit applied for Resend Email",
			}),
		)
	})
	handler.ServeHTTP(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	err := FromJSON(res.Body)

	if !errors.Is(err, ErrResourceExhausted) {
		t.Error("expected error to be ErrResourceExhausted")
	}

	var retErr RetriableError
	if !errors.As(err, &retErr) {
		t.Error("expected error to be RetriableError")
	}

	if got, want := retErr.GetRetryDelay(), time.Minute; got != want {
		t.Errorf("unexpected recommended retry delay; got %q, want %q", got, want)
	}

	var qfErr FailedQuotaError
	if !errors.As(err, &qfErr) {
		t.Error("expected error to be FailedQuotaError")
	}

	violation := qfErr.GetViolations()[0]
	if got, want := violation.GetDescription(), "Rate limit applied for Resend Email"; got != want {
		t.Errorf("unexpected quota violation description; got %q, want %q", got, want)

	}

	if got, want := violation.GetSubject(), "auth0|123456789"; got != want {
		t.Errorf("unexpected quota violation subject; got %q, want %q", got, want)
	}
}

type errFunc func(err error)

func (fn errFunc) Handle(err error) {
	fn(err)
}

func testHandler(t *testing.T) {
	t.Helper()
	SetErrorHandler(errFunc(func(err error) {
		t.Fatal(err)
	}))
}
