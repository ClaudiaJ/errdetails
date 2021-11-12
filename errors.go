package errdetails

import (
	"errors"
	"fmt"
	"time"

	"github.com/ClaudiaJ/errdetails/details"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
)

var errUnknown = errors.New("unknown error")

// Known Status Code errors for use as target of errors.Is().
//
// Prefer constructing new errors with New constructor.
var (
	ErrCanceled           error = &errCodeError{error: errUnknown, Code: codes.Canceled}
	ErrUnknown            error = &errCodeError{error: errUnknown, Code: codes.Unknown}
	ErrInvalidArgument    error = &errCodeError{error: errUnknown, Code: codes.InvalidArgument}
	ErrDeadlineExceeded   error = &errCodeError{error: errUnknown, Code: codes.DeadlineExceeded}
	ErrNotFound           error = &errCodeError{error: errUnknown, Code: codes.NotFound}
	ErrAlreadyExists      error = &errCodeError{error: errUnknown, Code: codes.AlreadyExists}
	ErrPermissionDenied   error = &errCodeError{error: errUnknown, Code: codes.PermissionDenied}
	ErrResourceExhausted  error = &errCodeError{error: errUnknown, Code: codes.ResourceExhausted}
	ErrFailedPrecondition error = &errCodeError{error: errUnknown, Code: codes.FailedPrecondition}
	ErrAborted            error = &errCodeError{error: errUnknown, Code: codes.Aborted}
	ErrOutOfRange         error = &errCodeError{error: errUnknown, Code: codes.OutOfRange}
	ErrUnimplemented      error = &errCodeError{error: errUnknown, Code: codes.Unimplemented}
	ErrInternal           error = &errCodeError{error: errUnknown, Code: codes.Internal}
	ErrUnavailable        error = &errCodeError{error: errUnknown, Code: codes.Unavailable}
	ErrDataLoss           error = &errCodeError{error: errUnknown, Code: codes.DataLoss}
	ErrUnauthenticated    error = &errCodeError{error: errUnknown, Code: codes.Unauthenticated}
)

// New creates a new error from an Status Error Code.
// Resulting errors can be checked with errors.Is to match exported implementations.
//
// This implementation is specific to gRPC Status codes, but the same example
// can be applied for any other Flag-like wrapping.
func New(code codes.Code, msg string, details ...Details) error {
	return WithDetails(&errCodeError{
		Code:  code,
		error: errors.New(msg),
	}, details...)
}

// errCodeError enriches an error with status codes.
type errCodeError struct {
	error
	codes.Code
}

// Error implements error, prefixes with the Status code string.
func (e *errCodeError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code.String(), e.error)
}

// Is implements errors.Is, matches a target error if it implements errorCode
// and the error code matches instance of errCodeError error code.
func (e *errCodeError) Is(target error) bool {
	if v, ok := target.(*errCodeError); ok {
		return v.Code == e.Code
	}

	return false
}

// GRPCStatus implements interface required for status.FromError to turn the
// error into a gRPC Status.
func (e *errCodeError) GRPCStatus() *status.Status {
	return status.New(e.Code, e.Error())
}

// Unwrap implements errors.Unwrap interface.
func (e *errCodeError) Unwrap() error {
	return e.error
}

// RequestInfoError is an error including metadata about the request that
// a client can attach when filing a bug or providing other forms of feedback.
type RequestInfoError interface {
	error
	details.RequestInfo
}

var _ RequestInfoError = (*errRequestInfo)(nil)

type errRequestInfo struct {
	error
	*errdetails.RequestInfo
}

// Unwrap implements errors.Unwrap interface.
func (e *errRequestInfo) Unwrap() error {
	return e.error
}

// HelpfulError is an error including links to documentation relevant to the
// error or API.
type HelpfulError interface {
	error
	WithLinks(...details.HelpLink) HelpfulError
	GetLinks() []details.HelpLink
}

type errHelpLink struct {
	error
	*errdetails.Help
}

// Unwrap implements errors.Unwrap interface.
func (h *errHelpLink) Unwrap() error {
	return h.error
}

// WithLinks adds links to relavant documentation.
func (h *errHelpLink) WithLinks(links ...details.HelpLink) HelpfulError {
	v := make([]*errdetails.Help_Link, len(links))

	for idx, link := range links {
		if helpLink, ok := link.(*errdetails.Help_Link); ok {
			v[idx] = helpLink
			continue
		}

		v[idx] = &errdetails.Help_Link{
			Description: link.GetDescription(),
			Url:         link.GetUrl(),
		}
	}

	h.Help.Links = append(h.Help.Links, v...)
	return h
}

// GetLinks retrieves all links annotated into the helpful error.
func (e *errHelpLink) GetLinks() []details.HelpLink {
	links := make([]details.HelpLink, len(e.Help.Links))
	for k, v := range e.Help.Links {
		links[k] = v
	}
	return links

}

// BadRequestError is an error indicating the client had made a bad request,
// and includes details of each violation of the field validation rules not
// satisfied by the request.
type BadRequestError interface {
	error
	WithViolation(violation ...details.FieldViolation) BadRequestError
	GetViolations() []details.FieldViolation
}

var _ BadRequestError = (*errBadRequest)(nil)

type errBadRequest struct {
	error
	*errdetails.BadRequest
}

// Unwrap implement errors.Unwrap
func (e *errBadRequest) Unwrap() error {
	return e.error
}

// WithViolation appends field violations to a bad request error.
func (e *errBadRequest) WithViolation(violations ...details.FieldViolation) BadRequestError {
	v := make([]*errdetails.BadRequest_FieldViolation, len(violations))

	for idx, fieldViolation := range violations {
		if violation, ok := fieldViolation.(*errdetails.BadRequest_FieldViolation); ok {
			v[idx] = violation
			continue
		}

		v[idx] = &errdetails.BadRequest_FieldViolation{
			Field:       fieldViolation.GetField(),
			Description: fieldViolation.GetDescription(),
		}
	}

	e.BadRequest.FieldViolations = append(e.BadRequest.FieldViolations, v...)
	return e
}

func (e *errBadRequest) GetViolations() []details.FieldViolation {
	violations := make([]details.FieldViolation, len(e.BadRequest.FieldViolations))
	for k, v := range e.BadRequest.FieldViolations {
		violations[k] = v
	}
	return violations
}

// DebugError is an error including debug information indicating where an error
// occurred and any additional details provided by the server.
type DebugError interface {
	error
	details.DebugInfo
}

var _ DebugError = (*errDebugInfo)(nil)

type errDebugInfo struct {
	error
	*errdetails.DebugInfo
}

// Unwrap implement errors.Unwrap
func (e *errDebugInfo) Unwrap() error {
	return e.error
}

// CausedError is an error describing the cause of an error with structured details.
type CausedError interface {
	error
	details.Info
}

var _ CausedError = (*errInfo)(nil)

type errInfo struct {
	error
	*errdetails.ErrorInfo
}

// Unwrap implements errors.Unwrap interface.
func (e *errInfo) Unwrap() error {
	return e.error
}

// LocalizedError is an error including a localized error message that is safe
// to return to the user.
type LocalizedError interface {
	error
	protoreflect.ProtoMessage
	details.LocalizedMessage
}

var _ LocalizedError = (*localizedError)(nil)

type localizedError struct {
	error
	*errdetails.LocalizedMessage
}

// Unwrap implements errors.Unwrap interface.
func (e *localizedError) Unwrap() error {
	return e.error
}

// FailedPreconditionError is an error describing what preconditions have failed.
//
// An example being a Terms of Service acknowledgement that may be required
// before using a particular API or service, responses from the service will
// indicate that the precondition has not been met.
type FailedPreconditionError interface {
	error
	WithViolation(...details.PreconditionViolation) FailedPreconditionError
	GetViolations() []details.PreconditionViolation
}

var _ FailedPreconditionError = (*errPreconditionFailed)(nil)

type errPreconditionFailed struct {
	error
	*errdetails.PreconditionFailure
}

// Unwrap implements errors.Unwrap interface.
func (e *errPreconditionFailed) Unwrap() error {
	return e.error
}

// WithViolation adds PreconditionViolations to the FailedPreconditionError.
func (e *errPreconditionFailed) WithViolation(violations ...details.PreconditionViolation) FailedPreconditionError {
	v := make([]*errdetails.PreconditionFailure_Violation, len(violations))

	for idx, fieldViolation := range violations {
		if violation, ok := fieldViolation.(*errdetails.PreconditionFailure_Violation); ok {
			v[idx] = violation
			continue
		}

		v[idx] = &errdetails.PreconditionFailure_Violation{
			Type:        fieldViolation.GetType(),
			Subject:     fieldViolation.GetSubject(),
			Description: fieldViolation.GetDescription(),
		}
	}

	e.PreconditionFailure.Violations = append(e.PreconditionFailure.Violations, v...)
	return e
}

// GetViolations gets all the PreconditionViolations on the FailedPreconditionError.
func (e *errPreconditionFailed) GetViolations() []details.PreconditionViolation {
	violations := make([]details.PreconditionViolation, len(e.PreconditionFailure.Violations))
	for k, v := range e.PreconditionFailure.Violations {
		violations[k] = v
	}
	return violations
}

// FailedQuotaError is an error describing a quota check failed.
type FailedQuotaError interface {
	error
	WithViolation(...details.QuotaViolation) FailedQuotaError
	GetViolations() []details.QuotaViolation
}

var _ FailedQuotaError = (*errQuotaFailure)(nil)

type errQuotaFailure struct {
	error
	*errdetails.QuotaFailure
}

// Unwrap implements errors.Unwrap interface.
func (e *errQuotaFailure) Unwrap() error {
	return e.error
}

// WithViolation adds quota violations to the FailedQuotaError.
func (e *errQuotaFailure) WithViolation(violations ...details.QuotaViolation) FailedQuotaError {
	v := make([]*errdetails.QuotaFailure_Violation, len(violations))

	for idx, fieldViolation := range violations {
		if violation, ok := fieldViolation.(*errdetails.QuotaFailure_Violation); ok {
			v[idx] = violation
			continue
		}

		v[idx] = &errdetails.QuotaFailure_Violation{
			Subject:     fieldViolation.GetSubject(),
			Description: fieldViolation.GetDescription(),
		}
	}

	e.QuotaFailure.Violations = append(e.QuotaFailure.Violations, v...)
	return e
}

// GetViolations gets all the QuotaViolations on the FailedQuotaError.
func (e *errQuotaFailure) GetViolations() []details.QuotaViolation {
	violations := make([]details.QuotaViolation, len(e.QuotaFailure.Violations))
	for k, v := range e.QuotaFailure.Violations {
		violations[k] = v
	}
	return violations
}

// ResourceInfoError is an error that describes the resource that is being accessed.
type ResourceInfoError interface {
	error
	details.ResourceInfo
}

var _ ResourceInfoError = (*errResourceInfo)(nil)

type errResourceInfo struct {
	error
	*errdetails.ResourceInfo
}

// Unwrap implements errors.Unwrap interface.
func (e *errResourceInfo) Unwrap() error {
	return e.error
}

// RetriableError is an error that describes when a client may retry a failed request.
//
// The retry delay represents a minimum duration in which the client is recommended to wait.
// It is always recommended the client should use exponential backoff when retrying.
type RetriableError interface {
	error
	WithDelay(time.Duration) RetriableError
	GetRetryDelay() time.Duration
}

var _ RetriableError = (*errRetryInfo)(nil)

type errRetryInfo struct {
	error
	*errdetails.RetryInfo
}

// Unwrap implements errors.Unwrap interface.
func (e *errRetryInfo) Unwrap() error {
	return e.error
}

// WithDelay sets a recommended retry delay on the RetriableError.
func (e *errRetryInfo) WithDelay(d time.Duration) RetriableError {
	e.RetryInfo.RetryDelay = durationpb.New(d)

	return e
}

// GetRetryDelay gets the recommended retry delay on the RetriableError.
func (e *errRetryInfo) GetRetryDelay() time.Duration {
	return e.RetryInfo.RetryDelay.AsDuration()
}
