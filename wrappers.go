package errdetails

import (
	"time"

	"github.com/ClaudiaJ/errdetails/details"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Details are just error wrappers.
type Details interface {
	Wrap(error) error
}

// WithDetails wrap an error with additional details.
func WithDetails(err error, details ...Details) error {
	for _, wrapper := range details {
		err = wrapper.Wrap(err)
	}

	return err
}

// wrapper implements ErrorWrapper
var _ Details = wrapperFunc(nil)

type wrapperFunc func(err error) error

// Wrap wraps an error inside another error
func (fn wrapperFunc) Wrap(err error) error {
	return fn(err)
}

// Code wraps an external error with a specified Status Code.
// Note that while it is possible to wrap an error with multiple status codes,
// only the outer layer will be considered the resulting Status Code when unwrapped.
func Code(code codes.Code) Details {
	return wrapperFunc(func(err error) error {
		return &errCodeError{error: err, Code: code}
	})
}

// BadRequest provides a Details wrapper to enrich errors with BadRequestError details.
func BadRequest(violations ...details.FieldViolation) Details {
	return wrapperFunc(func(err error) error {
		return WithBadRequest(err, violations...)
	})
}

// WithBadRequest wraps an error with Bad Request details having optional field violations.
func WithBadRequest(err error, violations ...details.FieldViolation) BadRequestError {
	v := make([]*errdetails.BadRequest_FieldViolation, 0, len(violations))
	req := &errBadRequest{
		error: err,
		BadRequest: &errdetails.BadRequest{
			FieldViolations: v,
		},
	}
	return req.WithViolation(violations...)
}

// Help wraps an error with links to documentation or FAQ pages.
func Help(links ...details.HelpLink) Details {
	return wrapperFunc(func(err error) error {
		return WithHelp(err, links...)
	})
}

// WithHelp provides a Details wrapper to enrich errors with HelpfulError details.
func WithHelp(err error, links ...details.HelpLink) HelpfulError {
	v := make([]*errdetails.Help_Link, 0, len(links))
	req := &errHelpLink{
		error: err,
		Help: &errdetails.Help{
			Links: v,
		},
	}
	return req.WithLinks(links...)
}

// RequestInfo provides a Details wrapper to enrich errors with RequestInfoError details.
func RequestInfo(info details.RequestInfo) Details {
	return wrapperFunc(func(err error) error {
		return WithRequestInfo(err, info)
	})
}

// WithRequestInfo adds RequestID and Serving Data to an error.
// Generally this is used to serve an ID that correlates with logging, along
// with encrypted stack trace or similar data relevant for serving a request.
// Errors then reported by testers or end users can be more readily triaged and
// troubleshot.
func WithRequestInfo(err error, info details.RequestInfo) RequestInfoError {
	var ok bool
	var details *errdetails.RequestInfo

	if details, ok = info.(*errdetails.RequestInfo); !ok {
		details = &errdetails.RequestInfo{
			RequestId:   info.GetRequestId(),
			ServingData: info.GetServingData(),
		}
	}
	return &errRequestInfo{error: err, RequestInfo: details}
}

// Debug provides a Details wrapper to enrich errors with DebugError details.
func Debug(info details.DebugInfo) Details {
	return wrapperFunc(func(err error) error {
		return WithDebug(err, info)
	})
}

// WithDebug wraps an error with additional debugging info.
func WithDebug(err error, info details.DebugInfo) DebugError {
	var ok bool
	var details *errdetails.DebugInfo

	if details, ok = info.(*errdetails.DebugInfo); !ok {
		details = &errdetails.DebugInfo{
			StackEntries: info.GetStackEntries(),
			Detail:       info.GetDetail(),
		}
	}

	return &errDebugInfo{error: err, DebugInfo: details}
}

// Cause provides a Details wrapper to enrich errors with CausedError details.
func Cause(info details.Info) Details {
	return wrapperFunc(func(err error) error {
		return WithCause(err, info)
	})
}

// WithCause wraps an error with information about the cause of the error.
func WithCause(err error, info details.Info) CausedError {
	var ok bool
	var details *errdetails.ErrorInfo

	if details, ok = info.(*errdetails.ErrorInfo); !ok {
		details = &errdetails.ErrorInfo{
			Reason:   info.GetReason(),
			Domain:   info.GetDomain(),
			Metadata: info.GetMetadata(),
		}
	}

	return &errInfo{error: err, ErrorInfo: details}
}

// LocalizedMessage provides a Details wrapper to enrich errors with LocalizedError details.
func LocalizedMessage(msg details.LocalizedMessage) Details {
	return wrapperFunc(func(err error) error {
		return WithLocalizedMessage(err, msg)
	})
}

// WithLocalizedMessage wraps an error providing a localized message that's safe to return to the end user.
func WithLocalizedMessage(err error, msg details.LocalizedMessage) LocalizedError {
	var ok bool
	var details *errdetails.LocalizedMessage

	if details, ok = msg.(*errdetails.LocalizedMessage); !ok {
		details = &errdetails.LocalizedMessage{
			Locale:  msg.GetLocale(),
			Message: msg.GetMessage(),
		}
	}

	return &localizedError{error: err, LocalizedMessage: details}
}

// PreconditionFailure provides a Details wrapper to enrich errors with FailedPreconditionError details.
func PreconditionFailure(violations ...details.PreconditionViolation) Details {
	return wrapperFunc(func(err error) error {
		return WithPreconditionFailure(err, violations...)
	})
}

// WithPreconditionFailure wraps an error describing what preconditions have failed to be met.
func WithPreconditionFailure(err error, violations ...details.PreconditionViolation) FailedPreconditionError {
	details := &errdetails.PreconditionFailure{
		Violations: make([]*errdetails.PreconditionFailure_Violation, 0, len(violations)),
	}
	wrapped := &errPreconditionFailed{error: err, PreconditionFailure: details}
	return wrapped.WithViolation(violations...)
}

// QuotaFailure provides a Details wrapper to enrich errors with FailedQuotaError details.
func QuotaFailure(violations ...details.QuotaViolation) Details {
	return wrapperFunc(func(err error) error {
		return WithQuotaFailure(err, violations...)
	})
}

// WithQuotaFailure wraps an error describing how a quota check has failed.
func WithQuotaFailure(err error, violations ...details.QuotaViolation) FailedQuotaError {
	details := &errdetails.QuotaFailure{
		Violations: make([]*errdetails.QuotaFailure_Violation, 0, len(violations)),
	}
	wrapped := &errQuotaFailure{error: err, QuotaFailure: details}
	return wrapped.WithViolation(violations...)
}

// RetryDelay provides a Details wrapper to enrich errors with RetriableError details.
func RetryDelay(delay time.Duration) Details {
	return wrapperFunc(func(err error) error {
		return WithRetryDelay(err, delay)
	})
}

// WithRetryDelay wraps an error indicating that a client may retry a failed
// request after a delay recommended here.
//
// It is always recommended that clients should use exponential backoff when
// retrying.
func WithRetryDelay(err error, delay time.Duration) RetriableError {

	return &errRetryInfo{
		error: err,
		RetryInfo: &errdetails.RetryInfo{
			RetryDelay: durationpb.New(delay),
		},
	}
}

// Resource provides a Details wrapper to enrich errors with ResourceInfoError details.
func Resource(info details.ResourceInfo) Details {
	return wrapperFunc(func(err error) error {
		return WithResource(err, info)
	})
}

// WithResource wraps an error with information about the resource that is being accessed.
func WithResource(err error, info details.ResourceInfo) ResourceInfoError {
	var ok bool
	var details *errdetails.ResourceInfo

	if details, ok = info.(*errdetails.ResourceInfo); !ok {
		details = &errdetails.ResourceInfo{
			ResourceType: info.GetResourceType(),
			ResourceName: info.GetResourceName(),
			Owner:        info.GetOwner(),
			Description:  info.GetDescription(),
		}
	}

	return &errResourceInfo{error: err, ResourceInfo: details}
}
