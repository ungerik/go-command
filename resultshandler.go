package command

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/ungerik/go-reflection"
)

type ResultsHandler interface {
	HandleResults(args Args, argVals, resultVals []reflect.Value, resultErr error) error
}

type ResultsHandlerFunc func(args Args, argVals, resultVals []reflect.Value, resultErr error) error

func (f ResultsHandlerFunc) HandleResults(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
	return f(args, argVals, resultVals, resultErr)
}

func resultsToInterfaces(results []reflect.Value) ([]interface{}, error) {
	r := make([]interface{}, len(results))
	for i, result := range results {
		resultInterface := result.Interface()

		if b, ok := resultInterface.([]byte); ok {
			r[i] = string(b)
			continue
		}

		switch reflection.DerefValue(result).Kind() {
		case reflect.Struct, reflect.Slice, reflect.Array:
			b, err := json.MarshalIndent(resultInterface, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("can't print command result as JSON because: %w", err)
			}
			r[i] = string(b)
			continue
		}

		r[i] = resultInterface
	}
	return r, nil
}

// PrintTo calls fmt.Fprint on writer with the result values as varidic arguments
func PrintTo(writer io.Writer) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
		if resultErr != nil {
			return resultErr
		}
		r, err := resultsToInterfaces(resultVals)
		if err != nil || len(r) == 0 {
			return err
		}
		_, err = fmt.Fprint(writer, r...)
		return err
	}
}

// PrintlnTo calls fmt.Fprintln on writer for every result
func PrintlnTo(writer io.Writer) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
		if resultErr != nil {
			return resultErr
		}
		results, err := resultsToInterfaces(resultVals)
		if err != nil || len(results) == 0 {
			return err
		}
		for _, r := range results {
			_, err = fmt.Fprintln(writer, r)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// Println calls fmt.Println for every result
var Println ResultsHandlerFunc = func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
	if resultErr != nil {
		return resultErr
	}
	results, err := resultsToInterfaces(resultVals)
	if err != nil || len(results) == 0 {
		return err
	}
	for _, r := range results {
		_, err = fmt.Println(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// PrintlnWithPrefixTo calls fmt.Fprintln(writer, prefix, result) for every result value
func PrintlnWithPrefixTo(prefix string, writer io.Writer) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
		if resultErr != nil {
			return resultErr
		}
		results, err := resultsToInterfaces(resultVals)
		if err != nil || len(results) == 0 {
			return err
		}
		for _, result := range results {
			_, err = fmt.Fprintln(writer, prefix, result)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// PrintlnWithPrefix calls fmt.Println(prefix, result) for every result value
func PrintlnWithPrefix(prefix string) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
		if resultErr != nil {
			return resultErr
		}
		results, err := resultsToInterfaces(resultVals)
		if err != nil || len(results) == 0 {
			return err
		}
		for _, result := range results {
			_, err = fmt.Println(prefix, result)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// Logger interface
type Logger interface {
	Printf(format string, args ...interface{})
}

// LogTo calls logger.Printf(fmt.Sprintln(results...))
func LogTo(logger Logger) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
		if resultErr != nil {
			return resultErr
		}
		results, err := resultsToInterfaces(resultVals)
		if err != nil || len(results) == 0 {
			return err
		}
		logger.Printf(fmt.Sprintln(results...))
		return nil
	}
}

// LogWithPrefixTo calls logger.Printf(fmt.Sprintln(results...)) with prefix prepended to the results
func LogWithPrefixTo(prefix string, logger Logger) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
		if resultErr != nil {
			return resultErr
		}
		results, err := resultsToInterfaces(resultVals)
		if err != nil || len(results) == 0 {
			return err
		}
		results = append([]interface{}{prefix}, results...)
		logger.Printf(fmt.Sprintln(results...))
		return nil
	}
}

// PrintlnText prints a fixed string if a command returns without an error
type PrintlnText string

func (t PrintlnText) HandleResults(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
	if resultErr != nil {
		return resultErr
	}
	_, err := fmt.Println(t)
	return err
}
