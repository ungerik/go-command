package command

import (
	"context"
	"reflect"
)

type Function interface {
	Name() string
	String() string

	ContextArg() bool
	NumArgs() int
	NumResults() int

	ArgNames() []string
	ArgTypes() []reflect.Type
	ResultTypes() []reflect.Type

	CallWithStrings(ctx context.Context, s ...string) (results []interface{}, err error)
	CallWithNamedStrings(ctx context.Context, m map[string]string) (results []interface{}, err error)
}
