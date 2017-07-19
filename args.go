package command

import (
	"bytes"
	"encoding"
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
	StringArgsFunc(commandFunc interface{}, argsDefOuterType reflect.Type, resultsHandlers []ResultsHandler) (StringArgsFunc, error)
	StringMapArgsFunc(commandFunc interface{}, argsDefOuterType reflect.Type, resultsHandlers []ResultsHandler) (StringMapArgsFunc, error)
	StringArgsResultValuesFunc(commandFunc interface{}, argsDefOuterType reflect.Type) (StringArgsResultValuesFunc, error)
	StringMapArgsResultValuesFunc(commandFunc interface{}, argsDefOuterType reflect.Type) (StringMapArgsResultValuesFunc, error)
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
	// err := args.(argsDefInner).init(commandFunc, reflect.TypeOf(args))
	// if err != nil {
	// 	return nil, err
	// }
	return args.StringArgsFunc(commandFunc, reflect.TypeOf(args), resultsHandlers)
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
	// err := args.(argsDefInner).init(commandFunc, reflect.TypeOf(args))
	// if err != nil {
	// 	return nil, err
	// }
	return args.StringMapArgsFunc(commandFunc, reflect.TypeOf(args), resultsHandlers)
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
	// err := args.(argsDefInner).init(commandFunc, reflect.TypeOf(args))
	// if err != nil {
	// 	return nil, err
	// }
	return args.StringArgsResultValuesFunc(commandFunc, reflect.TypeOf(args))
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
	// err := args.(argsDefInner).init(commandFunc, reflect.TypeOf(args))
	// if err != nil {
	// 	return nil, err
	// }
	return args.StringMapArgsResultValuesFunc(commandFunc, reflect.TypeOf(args))
}

func MustGetStringMapArgsResultValuesFunc(commandFunc interface{}, args Args) StringMapArgsResultValuesFunc {
	f, err := GetStringMapArgsResultValuesFunc(commandFunc, args)
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

func (def *ArgsDef) String() string {
	if !def.initialized {
		return "ArgsDef not initialized"
	}
	var buf bytes.Buffer
	for _, f := range def.argStructFields {
		if buf.Len() > 0 {
			buf.WriteByte(' ')
		}
		fmt.Fprintf(&buf, "<%s>", f.Name)
	}
	return buf.String()
}

func (def *ArgsDef) init(argsDefOuterType reflect.Type) {
	if def.initialized {
		return
	}
	def.outerType = reflection.DerefType(argsDefOuterType)
	def.argStructFields = reflection.FlatExportedStructFieldNames(def.outerType, "cmd")
	def.initialized = true
}

func (def *ArgsDef) checkFunctionSignature(commandFunc interface{}) (commandFuncVal reflect.Value, numArgs, errorIndex int, err error) {
	commandFuncVal = reflect.ValueOf(commandFunc)
	commandFuncType := commandFuncVal.Type()
	if commandFuncType.Kind() != reflect.Func {
		return reflect.Value{}, -1, -1, errors.Errorf("expected a function, but got %s", commandFuncType)
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
	argsStruct := reflect.New(def.outerType).Elem()
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
			return nil, err
		}
	}
	return argVals, nil
}

func (def *ArgsDef) StringArgsFunc(commandFunc interface{}, argsDefOuterType reflect.Type, resultsHandlers []ResultsHandler) (StringArgsFunc, error) {
	def.init(argsDefOuterType)

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
			err = resultsHandler.HandleResults(resultVals)
			if err != nil {
				return err
			}
		}

		return nil
	}, nil
}

func (def *ArgsDef) StringMapArgsFunc(commandFunc interface{}, argsDefOuterType reflect.Type, resultsHandlers []ResultsHandler) (StringMapArgsFunc, error) {
	def.init(argsDefOuterType)

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
			err = resultsHandler.HandleResults(resultVals)
			if err != nil {
				return err
			}
		}

		return nil
	}, nil
}

func (def *ArgsDef) StringArgsResultValuesFunc(commandFunc interface{}, argsDefOuterType reflect.Type) (StringArgsResultValuesFunc, error) {
	def.init(argsDefOuterType)

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

func (def *ArgsDef) StringMapArgsResultValuesFunc(commandFunc interface{}, argsDefOuterType reflect.Type) (StringMapArgsResultValuesFunc, error) {
	def.init(argsDefOuterType)

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
