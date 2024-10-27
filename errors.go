package duckron

type Error struct {
	operation string
	message   string
	rootErr   *error
}

func newError(operation string, message string) *Error {
	return &Error{
		operation: operation,
		message:   message,
	}
}

func (e *Error) wrap(err error) *Error {
	e.rootErr = &err
	return e
}

func (e *Error) unwrap() *error {
	return e.rootErr
}
