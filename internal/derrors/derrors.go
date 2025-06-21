package derrors

import "fmt"

// Wrap adds context to an error if the error is not nil.
// It wraps the error with additional context information.
func Wrap(errp *error, format string, args ...any) {
	if *errp != nil {
		*errp = fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), *errp)
	}
}
