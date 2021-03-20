package models

import "fmt"

// ErrNotFound is returned if a package is not installed.
type ErrNotFound struct {
	message string
}

// NewErrNotFound returns a pointer to a new instance of ErrNotFound.
func NewErrNotFound(message string, args ...interface{}) *ErrNotFound {
	return &ErrNotFound{
		message: fmt.Sprintf(message, args...),
	}
}

func (err *ErrNotFound) Error() string {
	return err.message
}
