package gogameloop

import (
	"fmt"
	"runtime/debug"
)

// LoopError is thrown when a gogameloop function returns an error.
type LoopError struct {
	Inner       error
	Message     string
	StackTrace  string
	ErrorSource TokenSource
	Misc        map[string]interface{}
}

func wrapLoopError(err error, source TokenSource, messagef string, msgArgs ...interface{}) LoopError {
	return LoopError{
		Inner:       err,
		Message:     fmt.Sprintf(messagef, msgArgs...),
		StackTrace:  string(debug.Stack()),
		ErrorSource: source,
		Misc:        make(map[string]interface{}),
	}
}

func (e LoopError) Error() string {
	return e.Message
}
