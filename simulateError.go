package gogameloop

import (
	"fmt"
	"runtime/debug"
)

// SimulateError is thrown when a Simulate function returns an error.
type SimulateError struct {
	Inner      error
	Message    string
	StackTrace string
	Misc       map[string]interface{}
}

func wrapSimulateError(err error, messagef string, msgArgs ...interface{}) SimulateError {
	return SimulateError{
		Inner:      err,
		Message:    fmt.Sprintf(messagef, msgArgs...),
		StackTrace: string(debug.Stack()),
		Misc:       make(map[string]interface{}),
	}
}

func (e SimulateError) Error() string {
	return e.Message
}
