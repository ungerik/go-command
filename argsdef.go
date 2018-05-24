package command

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	reflection "github.com/ungerik/go-reflection"
)

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
	var b strings.Builder
	for i, f := range def.argStructFields {
		if f.Name == "_" && i == len(def.argStructFields)-1 {
			// Don't show last field with name "_"
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "<%s:%s>", f.Name, reflection.DerefType(f.Field.Type))
	}
	return b.String()
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

func (def *ArgsDef) checkFunctionSignature(commandFunc interface{}) (commandFuncVal reflect.Value, numArgs int, varidic bool, errorIndex int, err error) {
	commandFuncVal = reflect.ValueOf(commandFunc)
	commandFuncType := commandFuncVal.Type()
	if commandFuncType.Kind() != reflect.Func {
		return reflect.Value{}, -1, false, -1, errors.Errorf("expected a function or method, but got %s", commandFuncType)
	}

	numResults := commandFuncType.NumOut()
	if numResults > 0 && commandFuncType.Out(numResults-1) == reflection.TypeOfError {
		errorIndex = numResults - 1
	} else {
		errorIndex = -1
	}

	numArgs = len(def.argStructFields)
	if numArgs != commandFuncType.NumIn() {
		return reflect.Value{}, -1, false, -1, errors.Errorf("number of fields in command.Args struct (%d) does not match number of function arguments (%d)", numArgs, commandFuncType.NumIn())
	}
	for i := range def.argStructFields {
		if def.argStructFields[i].Field.Type != commandFuncType.In(i) {
			return reflect.Value{}, -1, false, -1, errors.Errorf(
				"type of command.Args struct field '%s' is %s, which does not match function argument %d type %s",
				def.argStructFields[i].Field.Name,
				def.argStructFields[i].Field.Type,
				i,
				commandFuncType.In(i),
			)
		}
	}

	return commandFuncVal, numArgs, commandFuncType.IsVariadic(), errorIndex, nil
}

func (def *ArgsDef) argValsFromStringArgs(numFuncArgs int, callerArgs []string) ([]reflect.Value, error) {
	// Allocate a new args struct because we need addressable
	// variables of struct field types to hold arg values.
	// Instead of new individual variable use fields of args struct.
	argsStruct := reflect.New(def.outerStructType).Elem()
	argVals := make([]reflect.Value, numFuncArgs)
	numStringArgs := len(callerArgs)
	for i := range argVals {
		argVals[i] = argsStruct.FieldByIndex(def.argStructFields[i].Field.Index)
		if i >= numStringArgs {
			continue
		}
		err := assignString(argVals[i], callerArgs[i])
		if err != nil {
			return nil, err
		}
	}
	return argVals, nil
}

func (def *ArgsDef) argValsFromStringMapArgs(numFuncArgs int, callerArgs map[string]string) ([]reflect.Value, error) {
	// Allocate a new args struct because we need addressable
	// variables of struct field types to hold arg values.
	// Instead of new individual variable use fields of args struct.
	argsStruct := reflect.New(def.outerStructType).Elem()
	argVals := make([]reflect.Value, numFuncArgs)
	for i := range argVals {
		argVals[i] = argsStruct.FieldByIndex(def.argStructFields[i].Field.Index)
		argName := def.argStructFields[i].Name
		stringArg, hasArg := callerArgs[argName]
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

func (def *ArgsDef) argValsFromMapArgs(numFuncArgs int, callerArgs map[string]interface{}) ([]reflect.Value, error) {
	// Allocate a new args struct because we need addressable
	// variables of struct field types to hold arg values.
	// Instead of new individual variable use fields of args struct.
	argsStruct := reflect.New(def.outerStructType).Elem()
	argVals := make([]reflect.Value, numFuncArgs)
	for i := range argVals {
		argVals[i] = argsStruct.FieldByIndex(def.argStructFields[i].Field.Index)
		argName := def.argStructFields[i].Name
		varArg, hasArg := callerArgs[argName]
		if !hasArg {
			continue
		}
		err := assignAny(argVals[i], varArg)
		if err != nil {
			return nil, err
		}
	}
	return argVals, nil
}

func (def *ArgsDef) argValsFromJSON(numFuncArgs int, callerArgs []byte) ([]reflect.Value, error) {
	// Unmarshal callerArgs JSON to new args struct
	argsStructPtr := reflect.New(def.outerStructType)
	err := json.Unmarshal(callerArgs, argsStructPtr.Interface())
	if err != nil {
		return nil, err
	}

	argsStruct := argsStructPtr.Elem()
	argVals := make([]reflect.Value, numFuncArgs)
	for i := range argVals {
		argVals[i] = argsStruct.FieldByIndex(def.argStructFields[i].Field.Index)
	}
	return argVals, nil
}

func (def *ArgsDef) callFuncWithResultsHandlers(commandFuncVal reflect.Value, numArgs int, varidic bool, errorIndex int, argVals []reflect.Value, resultsHandlers []ResultsHandler) error {
	var resultVals []reflect.Value
	if varidic {
		resultVals = commandFuncVal.CallSlice(argVals)
	} else {
		resultVals = commandFuncVal.Call(argVals)
	}

	var resultErr error
	if errorIndex != -1 {
		resultErr, _ = resultVals[errorIndex].Interface().(error)
		resultVals = resultVals[:errorIndex]
	}
	for _, resultsHandler := range resultsHandlers {
		err := resultsHandler.HandleResults(def.outerArgs, argVals, resultVals, resultErr)
		if err != nil && err != resultErr {
			return err
		}
	}

	return resultErr
}

func (def *ArgsDef) callFuncAndReturnResults(commandFuncVal reflect.Value, numArgs int, varidic bool, errorIndex int, argVals []reflect.Value) ([]reflect.Value, error) {
	var resultVals []reflect.Value
	if varidic {
		resultVals = commandFuncVal.CallSlice(argVals)
	} else {
		resultVals = commandFuncVal.Call(argVals)
	}

	var resultErr error
	if errorIndex != -1 {
		resultErr, _ = resultVals[errorIndex].Interface().(error)
		resultVals = resultVals[:errorIndex]
	}
	return resultVals, resultErr
}

func (def *ArgsDef) StringArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (StringArgsFunc, error) {
	commandFuncVal, numFuncArgs, varidic, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(callerArgs ...string) error {
		argVals, err := def.argValsFromStringArgs(numFuncArgs, callerArgs)
		if err != nil {
			return err
		}
		return def.callFuncWithResultsHandlers(commandFuncVal, numFuncArgs, varidic, errorIndex, argVals, resultsHandlers)
	}, nil
}

func (def *ArgsDef) StringMapArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (StringMapArgsFunc, error) {
	commandFuncVal, numFuncArgs, varidic, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(callerArgs map[string]string) (err error) {
		argVals, err := def.argValsFromStringMapArgs(numFuncArgs, callerArgs)
		if err != nil {
			return err
		}
		return def.callFuncWithResultsHandlers(commandFuncVal, numFuncArgs, varidic, errorIndex, argVals, resultsHandlers)
	}, nil
}

func (def *ArgsDef) MapArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (MapArgsFunc, error) {
	commandFuncVal, numFuncArgs, varidic, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(callerArgs map[string]interface{}) (err error) {
		argVals, err := def.argValsFromMapArgs(numFuncArgs, callerArgs)
		if err != nil {
			return err
		}
		return def.callFuncWithResultsHandlers(commandFuncVal, numFuncArgs, varidic, errorIndex, argVals, resultsHandlers)
	}, nil
}

func (def *ArgsDef) JSONArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (JSONArgsFunc, error) {
	commandFuncVal, numFuncArgs, varidic, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(callerArgs []byte) (err error) {
		argVals, err := def.argValsFromJSON(numFuncArgs, callerArgs)
		if err != nil {
			return err
		}
		return def.callFuncWithResultsHandlers(commandFuncVal, numFuncArgs, varidic, errorIndex, argVals, resultsHandlers)
	}, nil
}

func (def *ArgsDef) StringArgsResultValuesFunc(commandFunc interface{}) (StringArgsResultValuesFunc, error) {
	commandFuncVal, numArgs, varidic, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(args []string) ([]reflect.Value, error) {
		argVals, err := def.argValsFromStringArgs(numArgs, args)
		if err != nil {
			return nil, err
		}
		return def.callFuncAndReturnResults(commandFuncVal, numArgs, varidic, errorIndex, argVals)
	}, nil
}

func (def *ArgsDef) StringMapArgsResultValuesFunc(commandFunc interface{}) (StringMapArgsResultValuesFunc, error) {
	commandFuncVal, numArgs, varidic, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(args map[string]string) ([]reflect.Value, error) {
		argVals, err := def.argValsFromStringMapArgs(numArgs, args)
		if err != nil {
			return nil, err
		}
		return def.callFuncAndReturnResults(commandFuncVal, numArgs, varidic, errorIndex, argVals)
	}, nil
}

func (def *ArgsDef) MapArgsResultValuesFunc(commandFunc interface{}) (MapArgsResultValuesFunc, error) {
	commandFuncVal, numArgs, varidic, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(args map[string]interface{}) ([]reflect.Value, error) {
		argVals, err := def.argValsFromMapArgs(numArgs, args)
		if err != nil {
			return nil, err
		}
		return def.callFuncAndReturnResults(commandFuncVal, numArgs, varidic, errorIndex, argVals)
	}, nil
}

func (def *ArgsDef) JSONArgsResultValuesFunc(commandFunc interface{}) (JSONArgsResultValuesFunc, error) {
	commandFuncVal, numArgs, varidic, errorIndex, err := def.checkFunctionSignature(commandFunc)
	if err != nil {
		return nil, err
	}

	return func(args []byte) ([]reflect.Value, error) {
		argVals, err := def.argValsFromJSON(numArgs, args)
		if err != nil {
			return nil, err
		}
		return def.callFuncAndReturnResults(commandFuncVal, numArgs, varidic, errorIndex, argVals)
	}, nil
}
