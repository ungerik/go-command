package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	reflection "github.com/ungerik/go-reflection"
)

var WithoutArgs ArgsDef

// ArgsDef implements Args
type ArgsDef struct {
	outerStructType reflect.Type
	argStructFields []reflection.NamedStructField
	argInfos        []Arg
	initialized     bool
}

func (def *ArgsDef) NumArgs() int {
	return len(def.argInfos)
}

func (def *ArgsDef) Args() []Arg {
	return def.argInfos
}

func (def *ArgsDef) ArgTag(index int, tag string) string {
	return def.argStructFields[index].Field.Tag.Get(tag)
}

// String implements the fmt.Stringer interface.
func (def *ArgsDef) String() string {
	if !def.initialized {
		return "ArgsDef not initialized"
	}
	var b strings.Builder
	for _, arg := range def.argInfos {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "<%s:%s>", arg.Name, reflection.DerefType(arg.Type))
	}
	return b.String()
}

// Init initializes ArgsDef with the reflection data from
// outerStructPtr wich has to be the address of the struct
// variable that embedds ArgsDef.
func (def *ArgsDef) Init(outerStructPtr interface{}) error {
	if def.initialized {
		return nil
	}

	if _, ok := outerStructPtr.(Args); !ok {
		return fmt.Errorf("outerStructPtr of type %T does not implement interface Args", outerStructPtr)
	}

	def.outerStructType = reflection.DerefType(reflect.TypeOf(outerStructPtr))
	if def.outerStructType.Kind() != reflect.Struct {
		return fmt.Errorf("ArgsDef must be contained in a struct, but outer type is %s", def.outerStructType)
	}

	def.argStructFields = reflection.FlatExportedNamedStructFields(def.outerStructType, ArgNameTag)

	def.argInfos = make([]Arg, len(def.argStructFields))
	for i := range def.argInfos {
		def.argInfos[i].Name = def.argStructFields[i].Name
		def.argInfos[i].Description = def.ArgTag(i, ArgDescriptionTag)
		def.argInfos[i].Type = def.argStructFields[i].Field.Type
	}

	def.initialized = true
	return nil
}

func (def *ArgsDef) argValsFromStringArgs(callerArgs []string) ([]reflect.Value, error) {
	// Allocate a new args struct because we need addressable
	// variables of struct field types to hold arg values.
	// Instead of new individual variable use fields of args struct.
	argsStruct := reflect.New(def.outerStructType).Elem()
	argVals := make([]reflect.Value, def.NumArgs())
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

func (def *ArgsDef) argValsFromStringMapArgs(callerArgs map[string]string) ([]reflect.Value, error) {
	// Allocate a new args struct because we need addressable
	// variables of struct field types to hold arg values.
	// Instead of new individual variable use fields of args struct.
	argsStruct := reflect.New(def.outerStructType).Elem()
	argVals := make([]reflect.Value, def.NumArgs())
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

func (def *ArgsDef) argValsFromMapArgs(callerArgs map[string]interface{}) ([]reflect.Value, error) {
	// Allocate a new args struct because we need addressable
	// variables of struct field types to hold arg values.
	// Instead of new individual variable use fields of args struct.
	argsStruct := reflect.New(def.outerStructType).Elem()
	argVals := make([]reflect.Value, def.NumArgs())
	for i := range argVals {
		argVals[i] = argsStruct.FieldByIndex(def.argStructFields[i].Field.Index)
		argName := def.argStructFields[i].Name
		varArg, hasArg := callerArgs[argName]
		if !hasArg {
			continue
		}
		err := assignAny(argVals[i], varArg) // TODO implement assignAny
		if err != nil {
			return nil, err
		}
	}
	return argVals, nil
}

func (def *ArgsDef) argValsFromJSON(argsJSON []byte) ([]reflect.Value, error) {
	argsJSON = bytes.TrimSpace(argsJSON)
	if len(argsJSON) < 2 {
		return nil, fmt.Errorf("invalid JSON: '%s'", string(argsJSON))
	}

	// Handle JSON array
	if argsJSON[0] == '[' {
		var callerArray []interface{}
		err := json.Unmarshal(argsJSON, &callerArray)
		if err != nil {
			return nil, err
		}
		argsStruct := reflect.New(def.outerStructType).Elem()
		argVals := make([]reflect.Value, def.NumArgs())
		for i := range argVals {
			argVals[i] = argsStruct.FieldByIndex(def.argStructFields[i].Field.Index)
			if i < len(callerArray) {
				err := assignAny(argVals[i], callerArray[i]) // TODO implement assignAny
				if err != nil {
					return nil, err
				}
			}
		}
		return argVals, nil
	}

	// Unmarshal argsJSON to new args struct
	argsStructPtr := reflect.New(def.outerStructType)
	err := json.Unmarshal(argsJSON, argsStructPtr.Interface())
	if err != nil {
		return nil, err
	}

	argsStruct := argsStructPtr.Elem()
	argVals := make([]reflect.Value, def.NumArgs())
	for i := range argVals {
		argVals[i] = argsStruct.FieldByIndex(def.argStructFields[i].Field.Index)
	}
	return argVals, nil
}

func (def *ArgsDef) StringArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (StringArgsFunc, error) {
	dispatcher, err := newFuncDispatcher(def, commandFunc)
	if err != nil {
		return nil, err
	}

	f := func(ctx context.Context, callerArgs ...string) error {
		argVals, err := def.argValsFromStringArgs(callerArgs)
		if err != nil {
			return err
		}
		return dispatcher.callWithResultsHandlers(ctx, argVals, resultsHandlers)
	}
	return f, nil
}

func (def *ArgsDef) StringMapArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (StringMapArgsFunc, error) {
	dispatcher, err := newFuncDispatcher(def, commandFunc)
	if err != nil {
		return nil, err
	}

	f := func(ctx context.Context, callerArgs map[string]string) (err error) {
		argVals, err := def.argValsFromStringMapArgs(callerArgs)
		if err != nil {
			return err
		}
		return dispatcher.callWithResultsHandlers(ctx, argVals, resultsHandlers)
	}
	return f, nil
}

func (def *ArgsDef) MapArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (MapArgsFunc, error) {
	dispatcher, err := newFuncDispatcher(def, commandFunc)
	if err != nil {
		return nil, err
	}

	f := func(ctx context.Context, callerArgs map[string]interface{}) (err error) {
		argVals, err := def.argValsFromMapArgs(callerArgs)
		if err != nil {
			return err
		}
		return dispatcher.callWithResultsHandlers(ctx, argVals, resultsHandlers)
	}
	return f, nil
}

func (def *ArgsDef) JSONArgsFunc(commandFunc interface{}, resultsHandlers []ResultsHandler) (JSONArgsFunc, error) {
	dispatcher, err := newFuncDispatcher(def, commandFunc)
	if err != nil {
		return nil, err
	}

	f := func(ctx context.Context, callerArgs []byte) (err error) {
		argVals, err := def.argValsFromJSON(callerArgs)
		if err != nil {
			return err
		}
		return dispatcher.callWithResultsHandlers(ctx, argVals, resultsHandlers)
	}
	return f, nil
}

func (def *ArgsDef) StringArgsResultValuesFunc(commandFunc interface{}) (StringArgsResultValuesFunc, error) {
	dispatcher, err := newFuncDispatcher(def, commandFunc)
	if err != nil {
		return nil, err
	}

	f := func(ctx context.Context, args []string) ([]reflect.Value, error) {
		argVals, err := def.argValsFromStringArgs(args)
		if err != nil {
			return nil, err
		}
		return dispatcher.callAndReturnResults(ctx, argVals)
	}
	return f, nil
}

func (def *ArgsDef) StringMapArgsResultValuesFunc(commandFunc interface{}) (StringMapArgsResultValuesFunc, error) {
	dispatcher, err := newFuncDispatcher(def, commandFunc)
	if err != nil {
		return nil, err
	}

	f := func(ctx context.Context, args map[string]string) ([]reflect.Value, error) {
		argVals, err := def.argValsFromStringMapArgs(args)
		if err != nil {
			return nil, err
		}
		return dispatcher.callAndReturnResults(ctx, argVals)
	}
	return f, nil
}

func (def *ArgsDef) MapArgsResultValuesFunc(commandFunc interface{}) (MapArgsResultValuesFunc, error) {
	dispatcher, err := newFuncDispatcher(def, commandFunc)
	if err != nil {
		return nil, err
	}

	f := func(ctx context.Context, args map[string]interface{}) ([]reflect.Value, error) {
		argVals, err := def.argValsFromMapArgs(args)
		if err != nil {
			return nil, err
		}
		return dispatcher.callAndReturnResults(ctx, argVals)
	}
	return f, nil
}

func (def *ArgsDef) JSONArgsResultValuesFunc(commandFunc interface{}) (JSONArgsResultValuesFunc, error) {
	dispatcher, err := newFuncDispatcher(def, commandFunc)
	if err != nil {
		return nil, err
	}

	f := func(ctx context.Context, argsJSON []byte) ([]reflect.Value, error) {
		argVals, err := def.argValsFromJSON(argsJSON)
		if err != nil {
			return nil, err
		}
		return dispatcher.callAndReturnResults(ctx, argVals)
	}
	return f, nil
}
