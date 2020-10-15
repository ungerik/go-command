package command

import (
	"context"
	"fmt"
	"reflect"
)

// type getReplacementValFunc func() reflect.Value

// var (
// 	ReplaceArgTypes = map[reflect.Type]getReplacementValFunc{
// 		typeOfContext: func() reflect.Value { return reflect.ValueOf(context.TODO()) },
// 	}
// )

// type argReplacement struct {
// 	argIndex          int  // index of the argument
// 	insert            bool // if the argument should be inserted or replaced
// 	getReplacementVal getReplacementValFunc
// }

// functionArgTypesWithoutReplaceables returns the function argument types except for
// the first argument of type context.Context and callback function arguments.
func functionArgTypesWithoutReplaceables(funcType reflect.Type) (argTypes []reflect.Type, firstArgIsContext bool, insertArgs []insertArg) {
	numArgs := funcType.NumIn()
	argTypes = make([]reflect.Type, 0, numArgs)
	for i := 0; i < numArgs; i++ {
		t := funcType.In(i)
		if i == 0 && t == typeOfContext {
			firstArgIsContext = true
			continue
		}
		if t.Kind() == reflect.Func {
			insertArgs = append(insertArgs, insertArg{index: i, value: reflect.Zero(t)})
			continue
		}
		// _, hasPlaceholder := ReplaceArgTypes[t]
		// if !hasPlaceholder {
		// 	argTypes = append(argTypes, t)
		// }
		argTypes = append(argTypes, t)
	}
	return argTypes, firstArgIsContext, insertArgs
}

type insertArg struct {
	index int
	value reflect.Value
}

type funcDispatcher struct {
	argsDef *ArgsDef

	funcVal  reflect.Value
	funcType reflect.Type

	// argReplacements []argReplacement

	firstArgIsContext bool
	insertArgs        []insertArg
	errorIndex        int
}

func newFuncDispatcher(argsDef *ArgsDef, commandFunc interface{}) (disp *funcDispatcher, err error) {
	disp = new(funcDispatcher)

	disp.argsDef = argsDef
	disp.funcVal = reflect.ValueOf(commandFunc)
	disp.funcType = disp.funcVal.Type()
	if disp.funcType.Kind() != reflect.Func {
		return nil, fmt.Errorf("expected a function or method, but got %s", disp.funcType)
	}

	numResults := disp.funcType.NumOut()
	if numResults > 0 && disp.funcType.Out(numResults-1) == typeOfError {
		disp.errorIndex = numResults - 1
	} else {
		disp.errorIndex = -1
	}

	// disp.argReplacements = nil // TODO

	var funcArgTypes []reflect.Type
	funcArgTypes, disp.firstArgIsContext, disp.insertArgs = functionArgTypesWithoutReplaceables(disp.funcType)
	numArgsDef := len(argsDef.argStructFields)
	if numArgsDef != len(funcArgTypes) {
		return nil, fmt.Errorf("number of fields in command.Args struct (%d) does not match number of function arguments (%d)", numArgsDef, len(funcArgTypes))
	}
	for i := range argsDef.argStructFields {
		if argsDef.argStructFields[i].Field.Type != funcArgTypes[i] {
			return nil, fmt.Errorf(
				"type of command.Args struct field '%s' is %s, which does not match function argument %d type %s",
				argsDef.argStructFields[i].Field.Name,
				argsDef.argStructFields[i].Field.Type,
				i,
				funcArgTypes[i],
			)
		}
	}

	return disp, nil
}

func (disp *funcDispatcher) callWithResultsHandlers(ctx context.Context, argVals []reflect.Value, resultsHandlers []ResultsHandler) error {
	if disp.firstArgIsContext {
		argVals = append([]reflect.Value{reflect.ValueOf(ctx)}, argVals...)
	}
	for _, insert := range disp.insertArgs {
		argVals = append(argVals[:insert.index], append([]reflect.Value{insert.value}, argVals[insert.index:]...)...)
	}

	var resultVals []reflect.Value
	if disp.funcType.IsVariadic() {
		resultVals = disp.funcVal.CallSlice(argVals)
	} else {
		resultVals = disp.funcVal.Call(argVals)
	}

	var resultErr error
	if disp.errorIndex != -1 {
		resultErr, _ = resultVals[disp.errorIndex].Interface().(error)
		resultVals = resultVals[:disp.errorIndex]
	}
	for _, resultsHandler := range resultsHandlers {
		err := resultsHandler.HandleResults(disp.argsDef, argVals, resultVals, resultErr)
		if err != nil && err != resultErr {
			return err
		}
	}

	return resultErr
}

func (disp *funcDispatcher) callAndReturnResults(ctx context.Context, argVals []reflect.Value) ([]reflect.Value, error) {
	if disp.firstArgIsContext {
		argVals = append([]reflect.Value{reflect.ValueOf(ctx)}, argVals...)
	}
	for _, insert := range disp.insertArgs {
		argVals = append(argVals[:insert.index], append([]reflect.Value{insert.value}, argVals[insert.index:]...)...)
	}

	var resultVals []reflect.Value
	if disp.funcType.IsVariadic() {
		resultVals = disp.funcVal.CallSlice(argVals)
	} else {
		resultVals = disp.funcVal.Call(argVals)
	}

	var resultErr error
	if disp.errorIndex != -1 {
		resultErr, _ = resultVals[disp.errorIndex].Interface().(error)
		resultVals = resultVals[:disp.errorIndex]
	}
	return resultVals, resultErr
}
