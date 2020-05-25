package global

import "fmt"

type Error interface {
	error
	Warning() bool
}

type globalError struct {
	msg       string
	isWarning bool
}

func (g globalError) Error() string {
	return g.msg
}

func (g globalError) Warning() bool {
	return g.isWarning
}

func NewWarning(msg string, arguments ...interface{}) Error {
	return &globalError{fmt.Sprintf(msg, arguments...), true}
}

func NewError(msg string, arguments ...interface{}) Error {
	return &globalError{fmt.Sprintf(msg, arguments...), false}
}
