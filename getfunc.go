package command

import (
	"context"
	"reflect"
)

type StringArgsFunc func(ctx context.Context, args ...string) error
type StringMapArgsFunc func(ctx context.Context, args map[string]string) error
type MapArgsFunc func(ctx context.Context, args map[string]interface{}) error
type JSONArgsFunc func(ctx context.Context, args []byte) error

type StringArgsResultValuesFunc func(ctx context.Context, args []string) ([]reflect.Value, error)
type StringMapArgsResultValuesFunc func(ctx context.Context, args map[string]string) ([]reflect.Value, error)
type MapArgsResultValuesFunc func(ctx context.Context, args map[string]interface{}) ([]reflect.Value, error)
type JSONArgsResultValuesFunc func(ctx context.Context, args []byte) ([]reflect.Value, error)

type argsImpl interface {
	Init(outerStructPtr interface{}) error

	StringArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (StringArgsFunc, error)
	StringMapArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (StringMapArgsFunc, error)
	MapArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (MapArgsFunc, error)
	JSONArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (JSONArgsFunc, error)

	StringArgsResultValuesFunc(commandFunc interface{}) (StringArgsResultValuesFunc, error)
	StringMapArgsResultValuesFunc(commandFunc interface{}) (StringMapArgsResultValuesFunc, error)
	MapArgsResultValuesFunc(commandFunc interface{}) (MapArgsResultValuesFunc, error)
	JSONArgsResultValuesFunc(commandFunc interface{}) (JSONArgsResultValuesFunc, error)
}

func GetStringArgsFunc(commandFunc interface{}, argsStructPtr interface{}, resultsHandlers ...ResultsHandler) (StringArgsFunc, error) {
	// Note: here happens something unexpected!
	// argsStructPtr is the address of the struct that embedds ArgsDef which in turn implements argsImpl
	// We need to pass the address of the outer args struct to ArgsDef.Init because ArgsDef doesn't
	// know anything about the struct it is embedded in.
	argsImpl := argsStructPtr.(argsImpl)
	err := argsImpl.Init(argsStructPtr)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringArgsFunc(commandFunc, resultsHandlers)
}

func MustGetStringArgsFunc(commandFunc interface{}, argsStructPtr interface{}, resultsHandlers ...ResultsHandler) StringArgsFunc {
	f, err := GetStringArgsFunc(commandFunc, argsStructPtr, resultsHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func GetStringMapArgsFunc(commandFunc interface{}, argsStructPtr interface{}, resultsHandlers ...ResultsHandler) (StringMapArgsFunc, error) {
	// Note: here happens something unexpected!
	// argsStructPtr is the address of the struct that embedds ArgsDef which in turn implements argsImpl
	// We need to pass the address of the outer args struct to ArgsDef.Init because ArgsDef doesn't
	// know anything about the struct it is embedded in.
	argsImpl := argsStructPtr.(argsImpl)
	err := argsImpl.Init(argsStructPtr)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringMapArgsFunc(commandFunc, resultsHandlers)
}

func MustGetStringMapArgsFunc(commandFunc interface{}, argsStructPtr interface{}, resultsHandlers ...ResultsHandler) StringMapArgsFunc {
	f, err := GetStringMapArgsFunc(commandFunc, argsStructPtr, resultsHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func GetMapArgsFunc(commandFunc interface{}, argsStructPtr interface{}, resultsHandlers ...ResultsHandler) (MapArgsFunc, error) {
	// Note: here happens something unexpected!
	// argsStructPtr is the address of the struct that embedds ArgsDef which in turn implements argsImpl
	// We need to pass the address of the outer args struct to ArgsDef.Init because ArgsDef doesn't
	// know anything about the struct it is embedded in.
	argsImpl := argsStructPtr.(argsImpl)
	err := argsImpl.Init(argsStructPtr)
	if err != nil {
		return nil, err
	}
	return argsImpl.MapArgsFunc(commandFunc, resultsHandlers)
}

func MustGetMapArgsFunc(commandFunc interface{}, argsStructPtr interface{}, resultsHandlers ...ResultsHandler) MapArgsFunc {
	f, err := GetMapArgsFunc(commandFunc, argsStructPtr, resultsHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func GetJSONArgsFunc(commandFunc interface{}, argsStructPtr interface{}, resultsHandlers ...ResultsHandler) (JSONArgsFunc, error) {
	// Note: here happens something unexpected!
	// argsStructPtr is the address of the struct that embedds ArgsDef which in turn implements argsImpl
	// We need to pass the address of the outer args struct to ArgsDef.Init because ArgsDef doesn't
	// know anything about the struct it is embedded in.
	argsImpl := argsStructPtr.(argsImpl)
	err := argsImpl.Init(argsStructPtr)
	if err != nil {
		return nil, err
	}
	return argsImpl.JSONArgsFunc(commandFunc, resultsHandlers)
}

func MustGetJSONArgsFunc(commandFunc interface{}, argsStructPtr interface{}, resultsHandlers ...ResultsHandler) JSONArgsFunc {
	f, err := GetJSONArgsFunc(commandFunc, argsStructPtr, resultsHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func GetStringArgsResultValuesFunc(commandFunc interface{}, argsStructPtr interface{}) (StringArgsResultValuesFunc, error) {
	// Note: here happens something unexpected!
	// argsStructPtr is the address of the struct that embedds ArgsDef which in turn implements argsImpl
	// We need to pass the address of the outer args struct to ArgsDef.Init because ArgsDef doesn't
	// know anything about the struct it is embedded in.
	argsImpl := argsStructPtr.(argsImpl)
	err := argsImpl.Init(argsStructPtr)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringArgsResultValuesFunc(commandFunc)
}

func MustGetStringArgsResultValuesFunc(commandFunc interface{}, argsStructPtr interface{}) StringArgsResultValuesFunc {
	f, err := GetStringArgsResultValuesFunc(commandFunc, argsStructPtr)
	if err != nil {
		panic(err)
	}
	return f
}

func GetStringMapArgsResultValuesFunc(commandFunc interface{}, argsStructPtr interface{}) (StringMapArgsResultValuesFunc, error) {
	// Note: here happens something unexpected!
	// argsStructPtr is the address of the struct that embedds ArgsDef which in turn implements argsImpl
	// We need to pass the address of the outer args struct to ArgsDef.Init because ArgsDef doesn't
	// know anything about the struct it is embedded in.
	argsImpl := argsStructPtr.(argsImpl)
	err := argsImpl.Init(argsStructPtr)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringMapArgsResultValuesFunc(commandFunc)
}

func MustGetStringMapArgsResultValuesFunc(commandFunc interface{}, argsStructPtr interface{}) StringMapArgsResultValuesFunc {
	f, err := GetStringMapArgsResultValuesFunc(commandFunc, argsStructPtr)
	if err != nil {
		panic(err)
	}
	return f
}

func GetMapArgsResultValuesFunc(commandFunc interface{}, argsStructPtr interface{}) (MapArgsResultValuesFunc, error) {
	// Note: here happens something unexpected!
	// argsStructPtr is the address of the struct that embedds ArgsDef which in turn implements argsImpl
	// We need to pass the address of the outer args struct to ArgsDef.Init because ArgsDef doesn't
	// know anything about the struct it is embedded in.
	argsImpl := argsStructPtr.(argsImpl)
	err := argsImpl.Init(argsStructPtr)
	if err != nil {
		return nil, err
	}
	return argsImpl.MapArgsResultValuesFunc(commandFunc)
}

func MustGetMapArgsResultValuesFunc(commandFunc interface{}, argsStructPtr interface{}) MapArgsResultValuesFunc {
	f, err := GetMapArgsResultValuesFunc(commandFunc, argsStructPtr)
	if err != nil {
		panic(err)
	}
	return f
}

func GetJSONArgsResultValuesFunc(commandFunc interface{}, argsStructPtr interface{}) (JSONArgsResultValuesFunc, error) {
	// Note: here happens something unexpected!
	// argsStructPtr is the address of the struct that embedds ArgsDef which in turn implements argsImpl
	// We need to pass the address of the outer args struct to ArgsDef.Init because ArgsDef doesn't
	// know anything about the struct it is embedded in.
	argsImpl := argsStructPtr.(argsImpl)
	err := argsImpl.Init(argsStructPtr)
	if err != nil {
		return nil, err
	}
	return argsImpl.JSONArgsResultValuesFunc(commandFunc)
}

func MustGetJSONArgsResultValuesFunc(commandFunc interface{}, argsStructPtr interface{}) JSONArgsResultValuesFunc {
	f, err := GetJSONArgsResultValuesFunc(commandFunc, argsStructPtr)
	if err != nil {
		panic(err)
	}
	return f
}
