package model

type RetryableError struct {
	error
}

func NewRetryableError(err error) RetryableError {
	return RetryableError{err}
}
