package errdetails_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ClaudiaJ/errdetails"
	detailspb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
)

var testErr error = errors.New("test error")

func TestCodeError(t *testing.T) {
	err := errdetails.New(codes.InvalidArgument, "bad request")

	if !errors.Is(err, errdetails.ErrInvalidArgument) {
		t.Error("errors.Is not ErrInvalidArgument")
	}
}

func TestBadRequestError(t *testing.T) {
	field, desc := "username", "username cannot be empty"
	err := errdetails.WithDetails(testErr,
		errdetails.BadRequest(&detailspb.BadRequest_FieldViolation{
			Field:       field,
			Description: desc,
		}))

	var badReq errdetails.BadRequestError
	if !errors.As(err, &badReq) {
		t.Error("errors.As not Bad Request error")
	}

	violation := badReq.GetViolations()[0]

	if got, want := violation.GetField(), field; got != want {
		t.Errorf("unexpected violation field; got %s, want %s", got, want)
	}
	if got, want := violation.GetDescription(), desc; got != want {
		t.Errorf("unexpected violation description; got %s, want %s", got, want)
	}
}

func TestRequestInfoError(t *testing.T) {
	id, data := "test ID", "test data"
	err := errdetails.WithDetails(testErr, errdetails.RequestInfo(&detailspb.RequestInfo{
		RequestId:   id,
		ServingData: data,
	}))

	var reqInfo errdetails.RequestInfoError
	if !errors.As(err, &reqInfo) {
		t.Error("errors.As not Request Info error")
	}

	if got, want := reqInfo.GetRequestId(), id; got != want {
		t.Errorf("unexpected request ID; got %s, want %s", got, want)
	}

	if got, want := reqInfo.GetServingData(), data; got != want {
		t.Errorf("unexpected serving data; got %s, want %s", got, want)
	}
}

func TestHelpfulError(t *testing.T) {
	url, desc := "https://exmaple.test/", "Example link"
	err := errdetails.WithDetails(testErr, errdetails.Help(&detailspb.Help_Link{
		Url:         url,
		Description: desc,
	}))

	var helpInfo errdetails.HelpfulError
	if !errors.As(err, &helpInfo) {
		t.Error("errors.As not Helpful error")
	}

	link := helpInfo.GetLinks()[0]

	if got, want := link.GetUrl(), url; got != want {
		t.Errorf("unexpected help link URL; got %s, want %s", got, want)
	}

	if got, want := link.GetDescription(), desc; got != want {
		t.Errorf("unexpected help link Description; got %s, want %s", got, want)
	}
}

func TestDebugError(t *testing.T) {
	err := errdetails.WithDetails(testErr, errdetails.Debug(&detailspb.DebugInfo{
		StackEntries: []string{},
		Detail:       "",
	}))

	var dbgErr errdetails.DebugError
	if !errors.As(err, &dbgErr) {
		t.Error("errors.As not Debug error")
	}
}

func TestCausedError(t *testing.T) {
	err := errdetails.WithDetails(testErr, errdetails.Cause(&detailspb.ErrorInfo{
		Reason:   "",
		Domain:   "",
		Metadata: map[string]string{},
	}))

	var info errdetails.CausedError
	if !errors.As(err, &info) {
		t.Error("errors.As not Info error")
	}
}

func TestLocalizedError(t *testing.T) {
	err := errdetails.WithDetails(testErr, errdetails.LocalizedMessage(&detailspb.LocalizedMessage{
		Locale:  "",
		Message: "",
	}))

	var localized errdetails.LocalizedError
	if !errors.As(err, &localized) {
		t.Error("errors.As not Localized error")
	}
}

func TestFailedPreconditionError(t *testing.T) {
	err := errdetails.WithDetails(testErr, errdetails.PreconditionFailure(&detailspb.PreconditionFailure_Violation{
		Type:        "",
		Subject:     "",
		Description: "",
	}))

	var condErr errdetails.FailedPreconditionError
	if !errors.As(err, &condErr) {
		t.Error("errors.As not Precondition error")
	}
}

func TestFailedQuotaError(t *testing.T) {
	err := errdetails.WithDetails(testErr, errdetails.QuotaFailure(&detailspb.QuotaFailure_Violation{
		Subject:     "",
		Description: "",
	}))

	var quoErr errdetails.FailedQuotaError
	if !errors.As(err, &quoErr) {
		t.Error("errors.As not Failed Quota error")
	}
}

func TestResourceInfoError(t *testing.T) {
	err := errdetails.WithDetails(testErr, errdetails.Resource(&detailspb.ResourceInfo{
		ResourceType: "",
		ResourceName: "",
		Owner:        "",
		Description:  "",
	}))

	var resErr errdetails.ResourceInfoError
	if !errors.As(err, &resErr) {
		t.Error("errors.As not ResourceInfoError")
	}
}

func TestRetriableError(t *testing.T) {
	delay := time.Second * 15
	err := errdetails.WithDetails(testErr, errdetails.RetryDelay(delay))

	var retErr errdetails.RetriableError
	if !errors.As(err, &retErr) {
		t.Error("errors.As not RetriableError")
	}

	if got, want := retErr.GetRetryDelay(), delay; got != want {
		t.Errorf("unexpected retry delay; got %v, want %v", got, want)
	}
}
