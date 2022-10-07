package utils

import (
	"fmt"
	"os"
)

// CheckForNilAndHandleError will check to see if an error is nil and handle
// the error if it exists.
func IfErrorExit(err error, message string) {
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", message, err)
		fmt.Println(wrappedErr)
		os.Exit(1)
	}
}

func ThrowError(message string) {
	err := fmt.Errorf("%s", message)
	fmt.Println(err)
	os.Exit(1)
}

// NewErrorMessage creates a new error with a given message.
func NewErrorMessage(message string, err error) error {
	return fmt.Errorf("%s: %w", message, err)
}
