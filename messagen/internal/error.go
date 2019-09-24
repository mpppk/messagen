package internal

import "errors"

type MessagenError interface {
	error
	Recoverable() bool
}

// errorString is a trivial implementation of error.
// implementation comes from errors package.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

type NoPickableTemplateError struct {
	error
}

func NewNoPickableTemplateError(text string) *NoPickableTemplateError {
	return &NoPickableTemplateError{error: errors.New("pickable template not found: " + text)}
}

func (e NoPickableTemplateError) Recoverable() bool {
	return true
}

type NoPickableDefinitionError struct {
	errorString
}

func NewNoPickableDefinitionError(text string) *NoPickableTemplateError {
	return &NoPickableTemplateError{error: errors.New("pickable definitin nt found" + text)}
}

func (e NoPickableDefinitionError) Recoverable() bool {
	return true
}
