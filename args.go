package command

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"

	"github.com/ungerik/go-reflection"
)

type StringArgsFunc func(args ...string) error
type StringMapArgsFunc func(args map[string]string) error

type Args interface {
	StringArgsFunc(argsDefType reflect.Type, commandFunc interface{}) (StringArgsFunc, error)
	StringMapArgsFunc(argsDefType reflect.Type, commandFunc interface{}) (StringMapArgsFunc, error)
}

func GetStringArgsFunc(args Args, commandFunc interface{}) (StringArgsFunc, error) {
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
	return args.StringArgsFunc(reflect.TypeOf(args), commandFunc)
}

func MustGetStringArgsFunc(args Args, commandFunc interface{}) StringArgsFunc {
	f, err := GetStringArgsFunc(args, commandFunc)
	if err != nil {
		panic(err)
	}
	return f
}

func GetStringMapArgsFunc(args Args, commandFunc interface{}) (StringMapArgsFunc, error) {
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
	return args.StringMapArgsFunc(reflect.TypeOf(args), commandFunc)
}

func MustGetStringMapArgsFunc(args Args, commandFunc interface{}) StringMapArgsFunc {
	f, err := GetStringMapArgsFunc(args, commandFunc)
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

type argsDefInner interface {
	init(argsDefType reflect.Type, commandFunc interface{}) error
}

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

func (def *ArgsDef) StringArgsFunc(argsDefOuterType reflect.Type, commandFunc interface{}) (StringArgsFunc, error) {
	def.init(argsDefOuterType)

	commandFuncVal := reflect.ValueOf(commandFunc)
	commandFuncType := commandFuncVal.Type()
	if commandFuncType.Kind() != reflect.Func {
		return nil, errors.New("not a function") // TODO better error desc
	}

	returnsError := commandFuncType.NumOut() == 1 && commandFuncType.Out(0) == reflection.TypeOfError
	returnsNothing := commandFuncType.NumOut() == 0
	if !returnsError && !returnsNothing {
		return nil, errors.New("not returning error") // TODO better error desc
	}

	numArgs := len(def.argStructFields)

	if numArgs != commandFuncType.NumIn() {
		return nil, errors.New("invalid arg num") // TODO better error desc
	}
	for i := range def.argStructFields {
		if def.argStructFields[i].Field.Type != commandFuncType.In(i) {
			return nil, errors.New("arg types not the same") // TODO better error desc
		}
	}

	f := func(stringArgs ...string) error {
		numStringArgs := len(stringArgs)
		newStruct := reflect.New(def.outerType).Elem()
		argVals := make([]reflect.Value, numArgs)
		for i := range argVals {
			argVals[i] = newStruct.FieldByIndex(def.argStructFields[i].Field.Index)
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

	return f, nil
}

func (def *ArgsDef) StringMapArgsFunc(argsDefOuterType reflect.Type, commandFunc interface{}) (StringMapArgsFunc, error) {
	def.init(argsDefOuterType)

	commandFuncVal := reflect.ValueOf(commandFunc)
	commandFuncType := commandFuncVal.Type()
	if commandFuncType.Kind() != reflect.Func {
		return nil, errors.New("not a function") // TODO better error desc
	}

	returnsError := commandFuncType.NumOut() == 1 && commandFuncType.Out(0) == reflection.TypeOfError
	returnsNothing := commandFuncType.NumOut() == 0
	if !returnsError && !returnsNothing {
		return nil, errors.New("not returning error") // TODO better error desc
	}

	numArgs := len(def.argStructFields)

	if numArgs != commandFuncType.NumIn() {
		return nil, errors.New("invalid arg num") // TODO better error desc
	}
	for i := range def.argStructFields {
		if def.argStructFields[i].Field.Type != commandFuncType.In(i) {
			return nil, errors.New("arg types not the same") // TODO better error desc
		}
	}

	f := func(args map[string]string) (err error) {
		newStruct := reflect.New(def.outerType).Elem()
		argVals := make([]reflect.Value, numArgs)
		for i := range argVals {
			argVals[i] = newStruct.FieldByIndex(def.argStructFields[i].Field.Index)
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
		if returnsError && resultVals[0].Interface() != nil {
			return resultVals[0].Interface().(error)
		}
		return nil
	}

	return f, nil
}
