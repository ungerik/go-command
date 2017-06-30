package command

import (
	"fmt"
	"reflect"

	"github.com/ungerik/go-reflection"
)

type StringArgsFunc func(args ...string) error

type Args interface {
	StringArgsFunc(argsDef, commandFunc interface{}) StringArgsFunc
}

func GetStringArgsFunc(args Args, commandFunc interface{}) StringArgsFunc {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But the first argument to the method has all the type information,
	// because here the complete outer embedding struct is passed.
	return args.StringArgsFunc(args, commandFunc)
}

type ArgsDef struct{}

func (*ArgsDef) StringArgsFunc(argsDef, commandFunc interface{}) StringArgsFunc {
	commandFuncValue := reflect.ValueOf(commandFunc)
	commandFuncType := commandFuncValue.Type()
	if commandFuncType.Kind() != reflect.Func {
		panic("not a function")
	}
	returnsError := commandFuncType.NumOut() == 1 && commandFuncType.Out(0) == reflection.TypeOfError
	returnsNothing := commandFuncType.NumOut() == 0
	if !returnsError && !returnsNothing {
		panic("not returning error")
	}

	numArgs := commandFuncType.NumIn()

	argDef := reflection.FlatStructFieldValueNames(argsDef, "cmd")
	if len(argDef) != numArgs {
		panic("invalid arg num")
	}
	for i := range argDef {
		if argDef[i].Type != commandFuncType.In(i) {
			panic("arg types not the same")
		}
	}

	return func(args ...string) error {
		numStringArgs := len(args)
		argVals := make([]reflect.Value, numArgs)
		for i := range argVals {
			// argVals[i] = reflect.Zero(argDef[i].Type)
			argVals[i] = argDef[i].Value
			if i < numStringArgs {
				if argDef[i].Type.Kind() == reflect.String {
					argVals[i].Set(reflect.ValueOf(args[i]))
				} else if argDef[i].Type.Kind() == reflect.Slice && argDef[i].Type.Elem().Kind() == reflect.Uint8 {
					argVals[i].Set(reflect.ValueOf([]byte(args[i])))
				} else {
					_, err := fmt.Sscan(args[i], argVals[i].Addr().Interface())
					if err != nil {
						return err
					}
				}
			}
		}

		resultVals := commandFuncValue.Call(argVals)
		if returnsError && resultVals[0].Interface() != nil {
			return resultVals[0].Interface().(error)
		}
		return nil
	}
}
