package errdetails_test

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/ClaudiaJ/errdetails"
	detailspb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
)

func ExampleNew() {
	// New creates a new error with a distinct Code
	err := errdetails.New(codes.InvalidArgument, "fields not satisfied")

	fmt.Println(errors.Is(err, errdetails.ErrInvalidArgument))
	//output:
	// true
}

func ExampleWithDetails() {
	// an error can be enriched with many additional sources of eror details
	errdetails.WithDetails(testErr,
		errdetails.BadRequest(),
		errdetails.RequestInfo(&detailspb.RequestInfo{
			RequestId: "123456789",
		}),
	)
}

func ExampleBadRequest() {
	errdetails.New(codes.InvalidArgument, "fields not satisfied",
		errdetails.BadRequest(
			// BadRequest takes optional Field Violations describing fields having failed validations.
			&detailspb.BadRequest_FieldViolation{
				Field:       "username",
				Description: "username must not contain any part of email address.",
			},
			&detailspb.BadRequest_FieldViolation{
				Field:       "password",
				Description: "password must be at least 5 characters long.",
			},
		),
	)
}

func ExampleCause() {
	errdetails.New(codes.DataLoss, "object payload not received in full",
		errdetails.Cause(&detailspb.ErrorInfo{
			Reason: "stream body prematurely terminated by client",
			Domain: "bucket.platform.test",
		}),
	)
}

func ExampleDebug() {
	errdetails.New(codes.Internal, "impossible error reached",
		errdetails.Debug(&detailspb.DebugInfo{
			StackEntries: []string{"data.Gnorm/One", "api.Thing/Something"},
			Detail:       "Request body was nil where it shouldn not have been",
		}),
	)
}

func ExampleHelp() {
	errdetails.New(codes.PermissionDenied, "access denied",
		errdetails.Help(&detailspb.Help_Link{
			Url:         "https://login.platform.test/",
			Description: "Login or Register to access this page.",
		}),
	)
}

func ExampleLocalizedMessage() {
	errdetails.New(codes.Unavailable, "service in maintenance mode",
		errdetails.LocalizedMessage(
			&detailspb.LocalizedMessage{
				Locale:  "en-US",
				Message: "Mattel Login is down for scheduled maintenance. Please try again later.",
			},
		),
	)
}

func ExamplePreconditionFailure() {
	errdetails.New(codes.FailedPrecondition, "Terms of Service is required",
		errdetails.PreconditionFailure(
			&detailspb.PreconditionFailure_Violation{
				Type:        "TOS",
				Description: "Please review and acknowledge Terms of Service before continuing.",
			},
		),
	)
}

func ExampleQuotaFailure() {
	errdetails.New(codes.ResourceExhausted, "Too many requests",
		errdetails.QuotaFailure(&detailspb.QuotaFailure_Violation{
			Description: "Rate limit exceeded.",
		}),
	)
}

func ExampleRequestInfo() {
	errdetails.New(codes.Internal, "unrecognized mime type on upload",
		errdetails.RequestInfo(&detailspb.RequestInfo{
			RequestId: "123456789",
		}),
	)
}

func ExampleResource() {
	errdetails.New(codes.ResourceExhausted, "no more Redlines available",
		errdetails.Resource(&detailspb.ResourceInfo{
			ResourceType: "currency",
			ResourceName: "Redlines",
			Owner:        "auth0|123456789",
			Description:  "no remaining currency in wallet",
		}),
	)
}

func ExampleRetryDelay() {
	errdetails.New(codes.Unavailable, "upstream responded with temporary failure", errdetails.RetryDelay(time.Minute))
}

func ExampleWithBadRequest() {
	err := errdetails.WithBadRequest(testErr,
		&detailspb.BadRequest_FieldViolation{
			Field:       "username",
			Description: "username must not contain any part of email address.",
		},
		&detailspb.BadRequest_FieldViolation{
			Field:       "password",
			Description: "password must be at least 5 characters long.",
		},
	)

	var badReq errdetails.BadRequestError
	if errors.As(err, &badReq) {
		fmt.Println("error is", reflect.ValueOf(&badReq).Elem().Type())
		for _, violation := range badReq.GetViolations() {
			fmt.Printf("field violation %q: %s\n", violation.GetField(), violation.GetDescription())
		}
	}
	//output:
	// error is errdetails.BadRequestError
	// field violation "username": username must not contain any part of email address.
	// field violation "password": password must be at least 5 characters long.
}

func ExampleWithCause() {
	const ReasonThrottle = "UPSTREAM_THROTTLE"

	err := errdetails.WithCause(testErr,
		&detailspb.ErrorInfo{
			Reason: ReasonThrottle,
			Domain: "fake.domain.test",
		},
	)

	var causedErr errdetails.CausedError
	if errors.As(err, &causedErr) {
		fmt.Println("error is", reflect.ValueOf(&causedErr).Elem().Type())
		fmt.Printf("with reason %q\n", causedErr.GetReason())
		fmt.Printf("with domain %q\n", causedErr.GetDomain())
	}
	//output:
	// error is errdetails.CausedError
	// with reason "UPSTREAM_THROTTLE"
	// with domain "fake.domain.test"
}

func ExampleWithDebug() {
	err := errdetails.WithDebug(testErr,
		&detailspb.DebugInfo{
			StackEntries: []string{"something", "goes", "here"},
			Detail:       "Server responded Internal Server Error with Message wrapping Status as string.",
		},
	)

	var debugErr errdetails.DebugError
	if errors.As(err, &debugErr) {
		fmt.Println("error is", reflect.ValueOf(&debugErr).Elem().Type())
		fmt.Printf("with stack entries: %q\n", debugErr.GetStackEntries())
		fmt.Printf("with detail: %q\n", debugErr.GetDetail())
	}
	//output:
	// error is errdetails.DebugError
	// with stack entries: ["something" "goes" "here"]
	// with detail: "Server responded Internal Server Error with Message wrapping Status as string."
}

func ExampleWithPreconditionFailure() {
	err := errdetails.WithPreconditionFailure(testErr, &detailspb.PreconditionFailure_Violation{
		Type:        "TOS",
		Description: "Terms of Service not accepted.",
	})

	var condErr errdetails.FailedPreconditionError
	if errors.As(err, &condErr) {
		fmt.Println("error is", reflect.ValueOf(&condErr).Elem().Type())
		for _, violation := range condErr.GetViolations() {
			fmt.Printf("precondition violation %q: %s\n", violation.GetType(), violation.GetDescription())
		}
	}
	//output:
	// error is errdetails.FailedPreconditionError
	// precondition violation "TOS": Terms of Service not accepted.
}

func ExampleWithQuotaFailure() {
	err := errdetails.WithQuotaFailure(testErr, &detailspb.QuotaFailure_Violation{
		Subject:     "auth0|123456789",
		Description: "Too many requests, too fast.",
	})

	var quotaErr errdetails.FailedQuotaError
	if errors.As(err, &quotaErr) {
		fmt.Println("error is", reflect.ValueOf(&quotaErr).Elem().Type())
		for _, violation := range quotaErr.GetViolations() {
			fmt.Printf("quota violation %q: %s\n", violation.GetSubject(), violation.GetDescription())
		}
	}
	//output:
	// error is errdetails.FailedQuotaError
	// quota violation "auth0|123456789": Too many requests, too fast.
}

func ExampleWithHelp() {
	err := errdetails.WithHelp(testErr, &detailspb.Help_Link{
		Url:         "https://wiki.platform.test/Why_Did_The_Error_Be",
		Description: "Article describing Platform standard errors and troubleshooting",
	})

	var helpErr errdetails.HelpfulError
	if errors.As(err, &helpErr) {
		fmt.Println("error is", reflect.ValueOf(&helpErr).Elem().Type())
		for _, link := range helpErr.GetLinks() {
			fmt.Printf("with url %q:\n%s\n", link.GetUrl(), link.GetDescription())
		}
	}
	//output:
	// error is errdetails.HelpfulError
	// with url "https://wiki.platform.test/Why_Did_The_Error_Be":
	// Article describing Platform standard errors and troubleshooting
}

func ExampleWithLocalizedMessage() {
	err := errdetails.WithLocalizedMessage(testErr, &detailspb.LocalizedMessage{
		Locale:  "es-MX",
		Message: "Enviar de nuevo",
	})

	var locErr errdetails.LocalizedError
	if errors.As(err, &locErr) {
		fmt.Println("error is", reflect.ValueOf(&locErr).Elem().Type())
		fmt.Printf("[%s]: %s\n", locErr.GetLocale(), locErr.GetMessage())
	}
	//output:
	// error is errdetails.LocalizedError
	// [es-MX]: Enviar de nuevo
}

func ExampleWithRequestInfo() {
	err := errdetails.WithRequestInfo(testErr, &detailspb.RequestInfo{
		RequestId: "123456789",
	})

	var reqErr errdetails.RequestInfoError
	if errors.As(err, &reqErr) {
		fmt.Println("error is", reflect.ValueOf(&reqErr).Elem().Type())
		fmt.Printf("with request ID %q\n", reqErr.GetRequestId())
	}
	//output:
	// error is errdetails.RequestInfoError
	// with request ID "123456789"
}

func ExampleWithResource() {
	err := errdetails.WithResource(testErr, &detailspb.ResourceInfo{
		ResourceType: "table",
		ResourceName: "public.shopify",
		Description:  "No record exists in table for shopify URL",
	})

	var resErr errdetails.ResourceInfoError
	if errors.As(err, &resErr) {
		fmt.Println("error is", reflect.ValueOf(&resErr).Elem().Type())
		fmt.Printf("with resource (%s) %q: %s\n", resErr.GetResourceType(), resErr.GetResourceName(), resErr.GetDescription())
	}
	//output:
	// error is errdetails.ResourceInfoError
	// with resource (table) "public.shopify": No record exists in table for shopify URL
}

func ExampleWithRetryDelay() {
	err := errdetails.WithRetryDelay(testErr, 10*time.Minute)

	var retErr errdetails.RetriableError
	if errors.As(err, &retErr) {
		fmt.Println("error is", reflect.ValueOf(&retErr).Elem().Type())
		fmt.Println("with recommended delay:", retErr.GetRetryDelay())
	}
	//output:
	// error is errdetails.RetriableError
	// with recommended delay: 10m0s
}
