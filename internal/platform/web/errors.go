package web

import "errors"

type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type ErrorResponse struct {
	Error  string       `json:"error"`
	Fields []FieldError `json:"fields,omitempty"`
}

type Error struct {
	Err    error
	Status int
	Fields []FieldError
}

func NewRequestError(err error, status int) error {
	return &Error{
		Err:    err,
		Status: status,
		Fields: nil,
	}
}

func (err *Error) Error() string {
	return err.Err.Error()
}

type shutdown struct {
	Message string
}

func (s *shutdown) Error() string {
	return s.Message
}

func NewShutdownError(message string) error {
	return &shutdown{message}
}

func IsShutdown(err error) bool {
	var s shutdown
	return errors.Is(err, &s)
}
