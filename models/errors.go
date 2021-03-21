package models

import "fmt"

// ErrNotFound is returned if a package is not installed or cannot be
// found in order to be installed.
type ErrNotFound struct {
	message string
}

// MakeErrNotFound returns a pointer to a new instance of ErrNotFound.
func MakeErrNotFound(message string, args ...interface{}) ErrNotFound {
	return ErrNotFound{
		message: fmt.Sprintf(message, args...),
	}
}

func (err ErrNotFound) Error() string {
	return err.message
}
