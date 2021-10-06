package command

import (
	"context"
	"reflect"
)

type Function interface {
	Name() string
	String() string

	NumArgs() int
	ContextArg() bool
	NumResults() int
	ErrorResult() bool

	ArgNames() []string
	ArgDescriptions() []string
	ArgTypes() []reflect.Type
	ResultTypes() []reflect.Type

	Call(ctx context.Context, args []interface{}) (results []interface{}, err error)
	CallWithStrings(ctx context.Context, strs ...string) (results []interface{}, err error)
	CallWithNamedStrings(ctx context.Context, strs map[string]string) (results []interface{}, err error)
}

func GenerateFunctionTODO(f interface{}) Function {
	panic("GenerateFunctionTODO: run gen-cmd-funcs")
}

// NewErrorFunction returns a Function that always
// returns the passed error when called.
func NewErrorFunction(err error) Function {
	return errFunction{err}
}

type errFunction struct {
	err error
}

func (e errFunction) Name() string                { return e.err.Error() }
func (e errFunction) String() string              { return e.err.Error() }
func (e errFunction) NumArgs() int                { return 0 }
func (e errFunction) ContextArg() bool            { return false }
func (e errFunction) NumResults() int             { return 0 }
func (e errFunction) ErrorResult() bool           { return false }
func (e errFunction) ArgNames() []string          { return nil }
func (e errFunction) ArgDescriptions() []string   { return nil }
func (e errFunction) ArgTypes() []reflect.Type    { return nil }
func (e errFunction) ResultTypes() []reflect.Type { return nil }

func (e errFunction) Call(ctx context.Context, args []interface{}) (results []interface{}, err error) {
	return nil, e.err
}

func (e errFunction) CallWithStrings(ctx context.Context, s ...string) (results []interface{}, err error) {
	return nil, e.err
}

func (e errFunction) CallWithNamedStrings(ctx context.Context, m map[string]string) (results []interface{}, err error) {
	return nil, e.err
}
