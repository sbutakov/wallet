package errors

import (
	"encoding/json"
	"runtime"
	"strconv"
)

var (
	// StackMaxDepth max depth for errors stack
	StackMaxDepth = 10
)

// F type for error fields
type F = map[string]interface{}

// New makes a new error
func New(msg string) error {
	return &withMessage{
		cause:  nil,
		msg:    msg,
		fields: make(F),
	}
}

// NewF makes a new error with fields
func NewF(msg string, fields F) error {
	return &withMessage{
		cause:  nil,
		msg:    msg,
		fields: fields,
	}
}

// Wrap wrap an error
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return WithMessage(err, msg)
}

// WrapF wrap an error and adds fields
func WrapF(err error, msg string, fields F) error {
	if err == nil {
		return nil
	}
	return WithFields(
		err,
		msg,
		fields,
	)
}

// WithMessage wrap an error with message
func WithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause:  err,
		msg:    msg,
		fields: make(map[string]interface{}),
	}
}

// WithFields wrap an error with message and fields
func WithFields(err error, msg string, fields map[string]interface{}) error {
	if err == nil {
		return nil
	}
	return &withMessage{
		cause:  err,
		msg:    msg,
		fields: fields,
	}
}

// WithCallers wrap an error with message, fields and callers
func WithCallers(err error, msg string, fields map[string]interface{}) error {
	if err == nil {
		return nil
	}
	pc := make([]uintptr, 32)
	n := runtime.Callers(2, pc)
	pc = pc[:n]
	callers := make([]string, 0, len(pc))
	frames := runtime.CallersFrames(pc)
	for {
		f, more := frames.Next()
		callers = append(callers, f.File+":"+strconv.Itoa(f.Line)+":"+f.Function)
		if !more {
			break
		}
	}
	return &withMessage{
		cause:   err,
		msg:     msg,
		fields:  fields,
		callers: callers,
	}
}

// Errors list of Error
type Errors []Error

// JSON returns list of errors encoded as JSON or error of marshalling as JSON string
func (e Errors) JSON() []byte {
	b, err := json.Marshal(e)
	if err != nil {
		return []byte(`"marshalling errors failed"`)
	}
	return b
}

// Error struct which represents an error
type Error struct {
	Msg     string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
	Callers []string               `json:"callers,omitempty"`
}

// Stack returns list of errors from all levels
func Stack(err error) Errors {
	if err == nil {
		return nil
	}

	errs := make(Errors, 0, StackMaxDepth)
	for i := 0; i < StackMaxDepth && err != nil; i++ {
		flds := Fields(err)
		errs = append(errs, Error{
			Msg:    err.Error(),
			Fields: flds,
		})
		err = Cause(err)
	}
	return errs
}

type withMessage struct {
	cause   error
	msg     string
	fields  F
	callers []string
}

func (w *withMessage) Error() string {
	return w.msg
}

func (w *withMessage) Cause() error {
	return w.cause
}

func (w *withMessage) Fields() map[string]interface{} {
	return w.fields
}

func (w *withMessage) Callers() []string {
	return w.callers
}

// Cause try to get cause from error
func Cause(err error) error {
	if e, ok := err.(causer); ok {
		return e.Cause()
	}
	return nil
}

// Fields try to get fields from error
func Fields(err error) map[string]interface{} {
	if e, ok := err.(fieldser); ok {
		flds := e.Fields()
		if len(flds) > 0 {
			return flds
		}
	}
	return nil
}

// Callers try to get callers from error
func Callers(err error) []string {
	if e, ok := err.(callers); ok {
		clrs := e.Callers()
		if len(clrs) > 0 {
			return clrs
		}
	}
	return nil
}

type callers interface {
	Callers() []string
}

type causer interface {
	Cause() error
}

type fieldser interface {
	Fields() map[string]interface{}
}
