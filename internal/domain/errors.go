package domain

import (
	"errors"
	"time"
)

var (
	ErrDescriptionTooLong = errors.New("description must be < 255 characters")
)

type Error struct {
	Code      string
	Message   string
	Details   string
	Timestamp time.Time
}

func (e *Error) Error() string {
	return e.Message
}

const (
	InvalidEntityCode = "INVALID"
	InternalCode      = "INTERNAL"
	ConflictCode      = "CONFLICT"
	NotFoundCode      = "NOT_FOUND"
)

type ErrOpts func(*Error) *Error

func NewError(code string, opts ...ErrOpts) *Error {
	err := &Error{
		Code: code,
	}

	for _, f := range opts {
		err = f(err)
	}

	return err
}

func NewErrorFrom(err error, opts ...ErrOpts) *Error {
	var e *Error
	if errors.As(err, &e) {
		for _, f := range opts {
			e = f(e)
		}
		return e
	}
	options := []ErrOpts{WithMessage(err.Error())}
	options = append(options, opts...)
	return NewError(InternalCode, options...)
}

func WithMessage(s string) ErrOpts {
	return func(err *Error) *Error {
		err.Message = s
		return err
	}
}

func WithDetails(s string) ErrOpts {
	return func(err *Error) *Error {
		err.Details = s
		return err
	}
}

func WithTS(ts time.Time) ErrOpts {
	return func(err *Error) *Error {
		err.Timestamp = ts
		return err
	}
}

// HasCode checks if an error has the specified error code
func HasCode(err error, code string) bool {
	var domainErr *Error
	if errors.As(err, &domainErr) {
		return domainErr.Code == code
	}
	return false
}
