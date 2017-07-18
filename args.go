package command

import (
	"encoding"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/ungerik/go-reflection"
)

type StringArgsFunc func(args ...string) error
type StringMapArgsFunc func(args map[string]string) error
type ResultHandlerFunc func(result reflect.Value) error

type Args interface {
	StringArgsFunc(argsDefType reflect.Type, commandFunc interface{}, resultHandlers []ResultHandlerFunc) (StringArgsFunc, error)
	StringMapArgsFunc(argsDefType reflect.Type, commandFunc interface{}, resultHandlers []ResultHandlerFunc) (StringMapArgsFunc, error)
}

func GetStringArgsFunc(args Args, commandFunc interface{}, resultHandlers ...ResultHandlerFunc) (StringArgsFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	// err := args.(argsDefInner).init(reflect.TypeOf(args), commandFunc)
	// if err != nil {
	// 	return nil, err
	// }
	return args.StringArgsFunc(reflect.TypeOf(args), commandFunc, resultHandlers)
}

func MustGetStringArgsFunc(args Args, commandFunc interface{}, resultHandlers ...ResultHandlerFunc) StringArgsFunc {
	f, err := GetStringArgsFunc(args, commandFunc, resultHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func GetStringMapArgsFunc(args Args, commandFunc interface{}, resultHandlers ...ResultHandlerFunc) (StringMapArgsFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	// err := args.(argsDefInner).init(reflect.TypeOf(args), commandFunc)
	// if err != nil {
	// 	return nil, err
	// }
	return args.StringMapArgsFunc(reflect.TypeOf(args), commandFunc, resultHandlers)
}

func MustGetStringMapArgsFunc(args Args, commandFunc interface{}, resultHandlers ...ResultHandlerFunc) StringMapArgsFunc {
	f, err := GetStringMapArgsFunc(args, commandFunc, resultHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func assignString(destVal reflect.Value, sourceStr string) (err error) {
	destPtr := destVal.Addr().Interface()
	switch v := destPtr.(type) {
	case encoding.TextUnmarshaler:
		err = v.UnmarshalText([]byte(sourceStr))
	case *[]byte:
		*v = []byte(sourceStr)
	default:
		// Don't check for type string directly,
		// use .Kind to match types derived from string
		if destVal.Kind() == reflect.String {
			destVal.Set(reflect.ValueOf(sourceStr))
		} else {
			// If all else fails, use fmt scanning
			// for generic type conversation from string
			_, err = fmt.Sscan(sourceStr, destPtr)
		}
	}
	return err
}

// type argsDefInner interface {
// 	init(argsDefType reflect.Type, commandFunc interface{}) error
// }

type ArgsDef struct {
	outerType       reflect.Type
	argStructFields []reflection.StructFieldName
	initialized     bool
}

func (def *ArgsDef) init(argsDefOuterType reflect.Type) {
	if def.initialized {
		return
	}
	def.outerType = reflection.DerefType(argsDefOuterType)
	def.argStructFields = reflection.FlatExportedStructFieldNames(def.outerType, "cmd")
	def.initialized = true
}

func (def *ArgsDef) checkFunctionSignature(commandFuncVal reflect.Value, resultHandlers []ResultHandlerFunc) (numArgs, errorIndex int, err error) {
	commandFuncType := commandFuncVal.Type()
	if commandFuncType.Kind() != reflect.Func {
		return -1, -1, errors.Errorf("expected a function, but got %s", commandFuncType)
	}

	numResults := commandFuncType.NumOut()
	switch numResults {
	case len(resultHandlers):
		// No error index in results
		errorIndex = -1

	case len(resultHandlers) + 1:
		// Last result is error
		errorIndex = numResults - 1
		if commandFuncType.Out(errorIndex) != reflection.TypeOfError {
			return -1, -1, errors.Errorf("expected error type as last result (index %d), but is %s", errorIndex, commandFuncType.Out(errorIndex))
		}

	default:
		return -1, -1, errors.Errorf("expected %d or %d results, but function has %d", len(resultHandlers), len(resultHandlers)+1, numResults)
	}

	numArgs = len(def.argStructFields)

	if numArgs != commandFuncType.NumIn() {
		return -1, -1, errors.Errorf("number of fields in command.Args struct (%d) does not match number of function arguments (%d)", numArgs, commandFuncType.NumIn())
	}
	for i := range def.argStructFields {
		if def.argStructFields[i].Field.Type != commandFuncType.In(i) {
			return -1, -1, errors.Errorf(
				"type of command.Args struct field '%s' is %s, which does not match function argument %d type %s",
				def.argStructFields[i].Field.Name,
				def.argStructFields[i].Field.Type,
				i,
				commandFuncType.In(i),
			)
		}
	}

	return numArgs, errorIndex, nil
}

func (def *ArgsDef) StringArgsFunc(argsDefOuterType reflect.Type, commandFunc interface{}, resultHandlers []ResultHandlerFunc) (StringArgsFunc, error) {
	def.init(argsDefOuterType)

	commandFuncVal := reflect.ValueOf(commandFunc)

	numArgs, errorIndex, err := def.checkFunctionSignature(commandFuncVal, resultHandlers)
	if err != nil {
		return nil, err
	}

	stringArgsFunc := func(stringArgs ...string) error {
		numStringArgs := len(stringArgs)
		// Allocate a new args struct because we need addressable
		// variables of struct field types to hold arg values.
		// Instead of new individual variable use fields of args struct.
		argsStruct := reflect.New(def.outerType).Elem()
		argVals := make([]reflect.Value, numArgs)
		for i := range argVals {
			argVals[i] = argsStruct.FieldByIndex(def.argStructFields[i].Field.Index)
			if i >= numStringArgs {
				continue
			}
			err := assignString(argVals[i], stringArgs[i])
			if err != nil {
				return err
			}
		}

		resultVals := commandFuncVal.Call(argVals)
		for i := range resultHandlers {
			err = resultHandlers[i](resultVals[i])
			if err != nil {
				return err
			}
		}
		if errorIndex != -1 && resultVals[errorIndex].Interface() != nil {
			return resultVals[0].Interface().(error)
		}
		return nil
	}

	return stringArgsFunc, nil
}

func (def *ArgsDef) StringMapArgsFunc(argsDefOuterType reflect.Type, commandFunc interface{}, resultHandlers []ResultHandlerFunc) (StringMapArgsFunc, error) {
	def.init(argsDefOuterType)

	commandFuncVal := reflect.ValueOf(commandFunc)

	numArgs, errorIndex, err := def.checkFunctionSignature(commandFuncVal, resultHandlers)
	if err != nil {
		return nil, err
	}

	stringMapArgsFunc := func(args map[string]string) (err error) {
		// Allocate a new args struct because we need addressable
		// variables of struct field types to hold arg values.
		// Instead of new individual variable use fields of args struct.
		argsStruct := reflect.New(def.outerType).Elem()
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
				return err
			}
		}

		resultVals := commandFuncVal.Call(argVals)
		for i := range resultHandlers {
			err = resultHandlers[i](resultVals[i])
			if err != nil {
				return err
			}
		}
		if errorIndex != -1 && resultVals[errorIndex].Interface() != nil {
			return resultVals[0].Interface().(error)
		}
		return nil
	}

	return stringMapArgsFunc, nil
}
