package details

// HelpLink describes a descriptive URL linking a user or tester to
// documentation relating to an error, API, or service.
type HelpLink interface {
	// GetUrl gets the URL of a link.
	GetUrl() string

	// GetDescription describes what the link offers.
	GetDescription() string
}

// FieldViolation represents a validated or required field that was evaluated
// to have not met requirement for the field.
type FieldViolation interface {
	// GetField describes the path to a field in the request body. The value
	// will be a sequence of dot-separated identifiers that identify a field.
	GetField() string

	// GetDescription describes why the request element is bad.
	GetDescription() string
}

// Metadata about the request that clients can attach when filling a bug or
// providing other forms of feedback.
type RequestInfo interface {
	// Opaque string that should only be interpreted by the service generating it.
	// For example, it could be used to identify requests in teh service's logs.
	GetRequestId() string

	// GetServingData holds additional data that was used to serve the request.
	// For example, an encrypted stack trace that can be sent back to the
	// service provider for debugging.
	GetServingData() string
}

// Info describes the cause of an error with structured details.
type Info interface {
	// GetReason gets the reason for the error.
	GetReason() string

	// GetDomain gets the logical grouping to which a "reason" belongs to.
	GetDomain() string

	// GetMetadata gets additional structured details about the error.
	GetMetadata() map[string]string
}

// PreconditionViolation describes a precondition that has failed resulting in an error.
type PreconditionViolation interface {
	// GetType gets the service-specific type of precondition failure.
	GetType() string

	// GetSubject gets the subject, relative to the type, that had failed.
	GetSubject() string

	// GetDescription gets the description of how th e precondition had failed.
	GetDescription() string
}

// QuotaViolation describes a single quota violation, for example a daily quota
// has been exceeded.
type QuotaViolation interface {
	// GetSubject gets the subject on which the quota check had failed.
	GetSubject() string

	// GetDescription gets a description of how the quota check had failed.
	GetDescription() string
}

// ResourceInfo describes a resource that is being accessed.
type ResourceInfo interface {
	// GetResourceType gets the name for the type of resource being accessed, e.g. "sql table",
	GetResourceType() string

	// GetResourceName gets the name of the resource being accessed, e.g. the name of a table in a database.
	GetResourceName() string

	// GetOwner gets the owner of a resource.
	GetOwner() string

	// GetDescription describes what error is encountered when accessing the resource.
	GetDescription() string
}

// DebugInfo describes additional debugging info.
type DebugInfo interface {
	// GetDetail gets additonal debugging information provided by the server.
	GetDetail() string

	// GetStackEntries gets stack entries indicating where the error occurred.
	GetStackEntries() []string
}

// LocalizedMessage provides a localized error message that is safe to return to the user.
type LocalizedMessage interface {
	// GetLocale gets the BCP 47 locale code for which the message is localized.
	GetLocale() string
	// GetMessage gets the localized message in the specified locale.
	GetMessage() string
}
