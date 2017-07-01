package command

import (
	"encoding"
	"fmt"
	"reflect"

	"github.com/ungerik/go-reflection"
)

type StringArgsFunc func(args ...string) error
type StringMapArgsFunc func(args map[string]string) error

type Args interface {
	StringArgsFunc(argsDefType reflect.Type, commandFunc interface{}) StringArgsFunc
	StringMapArgsFunc(argsDefType reflect.Type, commandFunc interface{}) StringMapArgsFunc
}

func GetStringArgsFunc(args Args, commandFunc interface{}) StringArgsFunc {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	return args.StringArgsFunc(reflect.TypeOf(args), commandFunc)
}

func GetStringMapArgsFunc(args Args, commandFunc interface{}) StringMapArgsFunc {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	return args.StringMapArgsFunc(reflect.TypeOf(args), commandFunc)
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

type ArgsDef struct{}

func (*ArgsDef) StringArgsFunc(argsDefType reflect.Type, commandFunc interface{}) StringArgsFunc {
	argsDefType = reflection.DerefType(argsDefType)

	commandFuncVal := reflect.ValueOf(commandFunc)
	commandFuncType := commandFuncVal.Type()
	if commandFuncType.Kind() != reflect.Func {
		panic("not a function")
	}

	returnsError := commandFuncType.NumOut() == 1 && commandFuncType.Out(0) == reflection.TypeOfError
	returnsNothing := commandFuncType.NumOut() == 0
	if !returnsError && !returnsNothing {
		panic("not returning error")
	}

	numArgs := commandFuncType.NumIn()

	argTypes := reflection.FlatStructFieldNames(argsDefType, "cmd")
	if len(argTypes) != numArgs {
		panic("invalid arg num")
	}
	for i := range argTypes {
		if argTypes[i].Field.Type != commandFuncType.In(i) {
			panic("arg types not the same")
		}
	}

	return func(stringArgs ...string) error {
		numStringArgs := len(stringArgs)
		argsDefVal := reflect.New(argsDefType).Elem()
		argVals := make([]reflect.Value, numArgs)
		for i := range argVals {
			argVals[i] = argsDefVal.FieldByIndex(argTypes[i].Field.Index)
			if i >= numStringArgs {
				continue
			}
			err := assignString(argVals[i], stringArgs[i])
			if err != nil {
				return err
			}
		}

		resultVals := commandFuncVal.Call(argVals)
		if returnsError && resultVals[0].Interface() != nil {
			return resultVals[0].Interface().(error)
		}
		return nil
	}
}

func (*ArgsDef) StringMapArgsFunc(argsDefType reflect.Type, commandFunc interface{}) StringMapArgsFunc {
	argsDefType = reflection.DerefType(argsDefType)

	commandFuncVal := reflect.ValueOf(commandFunc)
	commandFuncType := commandFuncVal.Type()
	if commandFuncType.Kind() != reflect.Func {
		panic("not a function")
	}

	returnsError := commandFuncType.NumOut() == 1 && commandFuncType.Out(0) == reflection.TypeOfError
	returnsNothing := commandFuncType.NumOut() == 0
	if !returnsError && !returnsNothing {
		panic("not returning error")
	}

	numArgs := commandFuncType.NumIn()

	argTypes := reflection.FlatStructFieldNames(argsDefType, "cmd")
	if len(argTypes) != numArgs {
		panic("invalid arg num")
	}
	for i := range argTypes {
		if argTypes[i].Field.Type != commandFuncType.In(i) {
			panic("arg types not the same")
		}
	}

	return func(args map[string]string) (err error) {
		argsDefVal := reflect.New(argsDefType).Elem()
		argVals := make([]reflect.Value, numArgs)
		for i := range argVals {
			argVals[i] = argsDefVal.FieldByIndex(argTypes[i].Field.Index)
			name := argTypes[i].Name
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
		if returnsError && resultVals[0].Interface() != nil {
			return resultVals[0].Interface().(error)
		}
		return nil
	}
}
