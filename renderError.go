package gogameloop

import (
	"fmt"
	"runtime/debug"
)

// RenderError is thrown when a Render function returns an error.
type RenderError struct {
	Inner      error
	Message    string
	StackTrace string
	Misc       map[string]interface{}
}

func wrapRenderError(err error, messagef string, msgArgs ...interface{}) RenderError {
	return RenderError{
		Inner:      err,
		Message:    fmt.Sprintf(messagef, msgArgs...),
		StackTrace: string(debug.Stack()),
		Misc:       make(map[string]interface{}),
	}
}

func (e RenderError) Error() string {
	return e.Message
}
