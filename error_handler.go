package errdetails

import "sync"

// ErrorHandler handles ierremediable events, e.g. to log error occurring while
// writing to http.ResponseWriter.
type ErrorHandler interface {
	Handle(error)
}

type errorHandler struct {
	mu      sync.RWMutex
	handler ErrorHandler
}

func (h *errorHandler) Handle(err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.handler != nil {
		h.handler.Handle(err)
	}
}

var (
	// handler may be used if set to handle or log errors that can't
	// be returned back to the caller.
	handler = &errorHandler{}
)

// SetLogger sets the package level logger that will be used when an error
// occurs after a Point of No Return and no error may be returned to caller
// (e.g. while writing to http.ResponseWriter)
//
// By default, these errors are thrown away.
func SetErrorHandler(h ErrorHandler) {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	handler.handler = h
}
