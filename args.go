package command

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/ungerik/go-reflection"
)

type StringArgsFunc func(args ...string) error
type StringMapArgsFunc func(args map[string]string) error

type StringArgsResultValuesFunc func(args []string) ([]reflect.Value, error)
type StringMapArgsResultValuesFunc func(args map[string]string) ([]reflect.Value, error)

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
	StringArgsResultValuesFunc(commandFunc interface{}) (StringArgsResultValuesFunc, error)
	StringMapArgsResultValuesFunc(commandFunc interface{}) (StringMapArgsResultValuesFunc, error)
}

func GetStringArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) (StringArgsFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringArgsFunc(commandFunc, resultsHandlers)
}

func MustGetStringArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) StringArgsFunc {
	f, err := GetStringArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func GetStringMapArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) (StringMapArgsFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringMapArgsFunc(commandFunc, resultsHandlers)
}

func MustGetStringMapArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) StringMapArgsFunc {
	f, err := GetStringMapArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func GetStringArgsResultValuesFunc(commandFunc interface{}, args Args) (StringArgsResultValuesFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringArgsResultValuesFunc(commandFunc)
}

func MustGetStringArgsResultValuesFunc(commandFunc interface{}, args Args) StringArgsResultValuesFunc {
	f, err := GetStringArgsResultValuesFunc(commandFunc, args)
	if err != nil {
		panic(err)
	}
	return f
}

func GetStringMapArgsResultValuesFunc(commandFunc interface{}, args Args) (StringMapArgsResultValuesFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringMapArgsResultValuesFunc(commandFunc)
}

func MustGetStringMapArgsResultValuesFunc(commandFunc interface{}, args Args) StringMapArgsResultValuesFunc {
	f, err := GetStringMapArgsResultValuesFunc(commandFunc, args)
	if err != nil {
		panic(err)
	}
	return f
}

var WithoutArgs ArgsDef

type ArgsDef struct {
	outerArgs       Args
	outerStructType reflect.Type
	argStructFields []reflection.StructFieldName
	initialized     bool
}

func (def *ArgsDef) NumArgs() int {
	return len(def.argStructFields)
}

func (def *ArgsDef) ArgName(index int) string {
	return def.argStructFields[index].Name
}

func (def *ArgsDef) ArgDescription(index int) string {
	return def.ArgTag(index, ArgDescriptionTag)
}

func (def *ArgsDef) ArgTag(index int, tag string) string {
	return def.argStructFields[index].Field.Tag.Get(tag)
}

func (def *ArgsDef) ArgType(index int) reflect.Type {
	return def.argStructFields[index].Field.Type
}

func (def *ArgsDef) String() string {
	if !def.initialized {
		return "ArgsDef not initialized"
	}
	var buf bytes.Buffer
	for _, f := range def.argStructFields {
		if buf.Len() > 0 {
			buf.WriteByte(' ')
		}
		fmt.Fprintf(&buf, "<%s:%s>", f.Name, f.Field.Type)
	}
	return buf.String()
}

func (def *ArgsDef) Init(outerArgs Args) error {
	if def.initialized {
		return nil
	}
	def.outerStructType = reflection.DerefType(reflect.TypeOf(outerArgs))
	if def.outerStructType.Kind() != reflect.Struct {
		return errors.Errorf("ArgsDef must be contained in a struct, but outer type is %s", def.outerStructType)
	}
	def.argStructFields = reflection.FlatExportedStructFieldNames(def.outerStructType, ArgNameTag)
	def.initialized = true
	return nil
}

func (def *ArgsDef) checkFunctionSignature(commandFunc interface{}) (commandFuncVal reflect.Value, numArgs, errorIndex int, err error) {
	commandFuncVal = reflect.ValueOf(commandFunc)
	commandFuncType := commandFuncVal.Type()
	if commandFuncType.Kind() != reflect.Func {
		return reflect.Value{}, -1, -1, errors.Errorf("expected a function or method, but got %s", commandFuncType)
	}

	numResults := commandFuncType.NumOut()
	if numResults > 0 && commandFuncType.Out(numResults-1) == reflection.TypeOfError {
		errorIndex = numResults - 1
	} else {
		errorIndex = -1
	}

	numArgs = len(def.argStructFields)
	if numArgs != commandFuncType.NumIn() {
		return reflect.Value{}, -1, -1, errors.Errorf("number of fields in command.Args struct (%d) does not match number of function arguments (%d)", numArgs, commandFuncType.NumIn())
	}
	for i := range def.argStructFields {
		if def.argStructFields[i].Field.Type != commandFuncType.In(i) {
			return reflect.Value{}, -1, -1, errors.Errorf(
				"type of command.Args struct field '%s' is %s, which does not match function argument %d type %s",
				def.argStructFields[i].Field.Name,
				def.argStructFields[i].Field.Type,
				i,
				commandFuncType.In(i),
			)
		}
	}

	return commandFuncVal, numArgs, errorIndex, nil
}

func (def *ArgsDef) getStringArgsVals(numArgs int, args []string) ([]reflect.Value, error) {
	// Allocate a new args struct because we need addressable
	// variables of struct field types to hold arg values.
	// Instead of new individual variable use fields of args struct.
	argsStruct := reflect.New(def.outerStructType).Elem()
	argVals := make([]reflect.Value, numArgs)
	numStringArgs := len(args)
	for i := range argVals {
		argVals[i] = argsStruct.FieldByIndex(def.argStructFields[i].Field.Index)
		if i >= numStringArgs {
			continue
		}
		err := assignString(argVals[i], args[i])
		if err != nil {
			return nil, err
		}
	}
	return argVals, nil
}

func (def *ArgsDef) getStringMapArgsVals(numArgs int, args map[string]string) ([]reflect.Value, error) {
	// Allocate a new args struct because we need addressable
	// variables of struct field types to hold arg values.
	// Instead of new individual variable use fields of args struct.
	argsStruct := reflect.New(def.outerStructType).Elem()
	argVals := make([]reflect.Value, numArgs)
	for i := range argVals {
		argVals[i] = argsStruct.FieldByIndex(def.argStructFields[i].Field.Index)
		name := def.argStructFields[i].Name
		stringArg, hasArg := args[name]
		if !hasArg {
			continue
		}
		err := assignString(argVals[i], stringArg)
		if err != nil {
			return nil, err
		}
	}
	return argVals, nil
}

func (def *ArgsDef) StringArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (StringArgsFunc, error) {
	commandFuncVal, numArgs, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(args ...string) error {
		argVals, err := def.getStringArgsVals(numArgs, args)
		if err != nil {
			return err
		}

		resultVals := commandFuncVal.Call(argVals)

		// Check for error first, don't handle other results if err != nil
		if errorIndex != -1 {
			err, _ = resultVals[errorIndex].Interface().(error)
			if err != nil {
				return err
			}
			resultVals = resultVals[:errorIndex]
		}
		for _, resultsHandler := range resultsHandlers {
			err = resultsHandler.HandleResults(def.outerArgs, argVals, resultVals)
			if err != nil {
				return err
			}
		}

		return nil
	}, nil
}

func (def *ArgsDef) StringMapArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (StringMapArgsFunc, error) {
	commandFuncVal, numArgs, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(args map[string]string) (err error) {
		argVals, err := def.getStringMapArgsVals(numArgs, args)
		if err != nil {
			return err
		}

		resultVals := commandFuncVal.Call(argVals)

		// Check for error first, don't handle other results if err != nil
		if errorIndex != -1 {
			err, _ = resultVals[errorIndex].Interface().(error)
			if err != nil {
				return err
			}
			resultVals = resultVals[:errorIndex]
		}
		for _, resultsHandler := range resultsHandlers {
			err = resultsHandler.HandleResults(def.outerArgs, argVals, resultVals)
			if err != nil {
				return err
			}
		}

		return nil
	}, nil
}

func (def *ArgsDef) StringArgsResultValuesFunc(commandFunc interface{}) (StringArgsResultValuesFunc, error) {
	commandFuncVal, numArgs, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(args []string) ([]reflect.Value, error) {
		argVals, err := def.getStringArgsVals(numArgs, args)
		if err != nil {
			return nil, err
		}

		resultVals := commandFuncVal.Call(argVals)

		// Check for error first, don't handle other results if err != nil
		if errorIndex != -1 {
			err, _ = resultVals[errorIndex].Interface().(error)
			if err != nil {
				return nil, err
			}
			resultVals = resultVals[:errorIndex]
		}
		return resultVals, nil
	}, nil
}

func (def *ArgsDef) StringMapArgsResultValuesFunc(commandFunc interface{}) (StringMapArgsResultValuesFunc, error) {
	commandFuncVal, numArgs, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(args map[string]string) ([]reflect.Value, error) {
		argVals, err := def.getStringMapArgsVals(numArgs, args)
		if err != nil {
			return nil, err
		}

		resultVals := commandFuncVal.Call(argVals)

		// Check for error first, don't handle other results if err != nil
		if errorIndex != -1 {
			err, _ = resultVals[errorIndex].Interface().(error)
			if err != nil {
				return nil, err
			}
			resultVals = resultVals[:errorIndex]
		}
		return resultVals, nil
	}, nil
}
