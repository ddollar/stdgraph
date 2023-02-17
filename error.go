package stdgraph

import "fmt"

type Error interface {
	Code() int
	Error() string
}

type httpError struct {
	error
	code int
}

func (he httpError) Code() int {
	return he.code
}

func (he httpError) Error() string {
	return he.error.Error()
}

func Errorf(code int, format string, args ...interface{}) Error {
	return httpError{
		error: fmt.Errorf(format, args...),
		code:  code,
	}
}
