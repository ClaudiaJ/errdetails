# Error Detail wrappers

This project provides wrappers around [`googleapis/rpc/errdetails`][#1] types,
as well as an error type wrapping gRPC status codes.

An error created or wrapped by this package can be unwrapped and transcribed to
a Status message type with all the added details intact.

Until inspiration strikes again, all the errors contained herein are derivative
of Google's `errdetails` messages.

Consider unwrapping these errors to enrich logs as well, they're genuinely
useful constructs for everyday purposes.

## Why, though?

Story time.

We never utilized the `WithDetails` functionality of gRPC Status messages, or
even `*status.Status` at all, preferring instead to use regular old errors with
a translation layer in middleware.

There's so much more knowledge you can express in a Status response coming
right from your API layer though, and I wished for some way to make the Status
message and arbitrary details more palatable to a team already familiar with
error wrapping.

At the same time, I wanted a unified error response body from all our JSON
API's, even the ones not transcoded from gRPC.

If only I could just, wrap a Status message with Details like an error.

And if only I could write my wrapped error right through a middleware layer to
turn it into JSON.

Fast forward.

Until recently, I was struggling to understand when I should use an interface
type in Go.
I understood `what`, and mostly `why`, but the intuition wasn't all here.

While working on localizations for a project, it started to click a bit.
Inspiration hit when I wanted to combine localizability with errors, and maybe
also a request ID, and wouldn't ya know it suddenly I couldn't sleep with all
the notes I had to write to myself in the "Probably Ideas" category, because
`errdetails` is a thing that exists.

So to reinforce and practice the point till intuition, and also to get it out
of the head so I can sleep again, I popped this old idea off the ol' `// TODO:`
list and got to writing something I'll probably never maintain or use, but
maybe I will and it'll be nice.

[#1]: https://pkg.go.dev/google.golang.org/genproto/googleapis/rpc/errdetails
