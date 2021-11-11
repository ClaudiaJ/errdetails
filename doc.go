// Package errdetails provides a convenient wrapping mechanism to incorporate
// gRPC Status details with Go's error wrapping paradigm.
//
// Errors provided by this package are implemented by embedding protobuf
// errdetails messages, and themselves implement an identical interface.
// This allows the use of protobuf errdetails messages in the method signature
// of each of the Detail wrappers, any custom implementation thereof, or even
// another error unwrapped by `errors.As`.
//
// Each error interface can be used in errors.As or errors.Is functions to
// unwrap, and some enable appending further details this way.
package errdetails
