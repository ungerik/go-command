package command

import (
	"reflect"
)

type StringArgsFunc func(args ...string) error
type StringMapArgsFunc func(args map[string]string) error
type MapArgsFunc func(args map[string]interface{}) error
type JSONArgsFunc func(args []byte) error

type StringArgsResultValuesFunc func(args []string) ([]reflect.Value, error)
type StringMapArgsResultValuesFunc func(args map[string]string) ([]reflect.Value, error)
type MapArgsResultValuesFunc func(args map[string]interface{}) ([]reflect.Value, error)
type JSONArgsResultValuesFunc func(args []byte) ([]reflect.Value, error)

type Args interface {
	NumArgs() int
	ArgName(index int) string
	ArgDescription(index int) string
	ArgTag(index int, tag string) string
	ArgType(index int) reflect.Type
	String() string
}

type ArgsImpl interface {
	Init(outerArgs Args) error
	StringArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (StringArgsFunc, error)
	StringMapArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (StringMapArgsFunc, error)
	MapArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (MapArgsFunc, error)
	JSONArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (JSONArgsFunc, error)
	StringArgsResultValuesFunc(commandFunc interface{}) (StringArgsResultValuesFunc, error)
	StringMapArgsResultValuesFunc(commandFunc interface{}) (StringMapArgsResultValuesFunc, error)
	MapArgsResultValuesFunc(commandFunc interface{}) (MapArgsResultValuesFunc, error)
	JSONArgsResultValuesFunc(commandFunc interface{}) (JSONArgsResultValuesFunc, error)
}
