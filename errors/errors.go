package errors

import "fmt"

// ErrAPICallFailed adalah error ketika panggilan ke API eksternal gagal.
type ErrAPICallFailed struct {
	StatusCode int
	Message    string
}

func (e *ErrAPICallFailed) Error() string {
	return fmt.Sprintf("API call failed with status %d: %s", e.StatusCode, e.Message)
}

// ErrDBOperationFailed adalah error ketika operasi database gagal.
type ErrDBOperationFailed struct {
	Operation string
	Err       error
}

func (e *ErrDBOperationFailed) Error() string {
	return fmt.Sprintf("database operation '%s' failed: %v", e.Operation, e.Err)
}

func (e *ErrDBOperationFailed) Unwrap() error {
	return e.Err
}
