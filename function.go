package command

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type FunctionInfo interface {
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
}

type Function interface {
	FunctionInfo

	Call(ctx context.Context, args []interface{}) (results []interface{}, err error)
	CallWithStrings(ctx context.Context, strs ...string) (results []interface{}, err error)
	CallWithNamedStrings(ctx context.Context, strs map[string]string) (results []interface{}, err error)
}

func GenerateFunctionTODO(f interface{}) Function {
	panic("GenerateFunctionTODO: run gen-cmd-funcs")
}

func ReflectFunctionInfo(name string, f interface{}) (FunctionInfo, error) {
	t := reflect.ValueOf(f).Type()
	if t.Kind() != reflect.Func {
		return nil, fmt.Errorf("%s passed instead of a function", t)
	}
	info := &functionInfo{
		name:            name,
		argNames:        make([]string, t.NumIn()),
		argDescriptions: make([]string, t.NumIn()),
		argTypes:        make([]reflect.Type, t.NumIn()),
		resultTypes:     make([]reflect.Type, t.NumOut()),
	}
	for i := range info.argTypes {
		info.argNames[i] = "a" + strconv.Itoa(i)
		info.argTypes[i] = t.In(i)
	}
	for i := range info.resultTypes {
		info.resultTypes[i] = t.Out(i)
	}
	return info, nil
}

type functionInfo struct {
	name            string
	argNames        []string
	argDescriptions []string
	argTypes        []reflect.Type
	resultTypes     []reflect.Type
}

func (f *functionInfo) Name() string { return f.name }
func (f *functionInfo) String() string {
	var b strings.Builder
	b.WriteString(f.name)
	b.WriteByte('(')
	for i, argName := range f.argNames {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(argName)
		b.WriteByte(' ')
		b.WriteString(f.argTypes[i].String())
	}
	b.WriteByte(')')
	return b.String()
}
func (f *functionInfo) NumArgs() int { return len(f.argNames) }
func (f *functionInfo) ContextArg() bool {
	return len(f.argTypes) > 0 && f.argTypes[0].String() == "context.Context"
}
func (f *functionInfo) NumResults() int { return len(f.resultTypes) }
func (f *functionInfo) ErrorResult() bool {
	return len(f.resultTypes) > 0 && f.resultTypes[0].String() == "error"
}
func (f *functionInfo) ArgNames() []string          { return f.argNames }
func (f *functionInfo) ArgDescriptions() []string   { return f.argDescriptions }
func (f *functionInfo) ArgTypes() []reflect.Type    { return f.argTypes }
func (f *functionInfo) ResultTypes() []reflect.Type { return f.resultTypes }

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
