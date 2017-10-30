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
				return nil, err
			}
			r[i] = string(b)
			continue
		}

		r[i] = resultInterface
	}
	return r, nil
}

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

func PrintlnTo(writer io.Writer) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
		if resultErr != nil {
			return resultErr
		}
		r, err := resultsToInterfaces(resultVals)
		if err != nil || len(r) == 0 {
			return err
		}
		_, err = fmt.Fprintln(writer, r...)
		return err
	}
}

var Println ResultsHandlerFunc = func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
	if resultErr != nil {
		return resultErr
	}
	r, err := resultsToInterfaces(resultVals)
	if err != nil || len(r) == 0 {
		return err
	}
	_, err = fmt.Println(r...)
	return err
}

func PrintlnWithPrefixTo(prefix string, writer io.Writer) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
		if resultErr != nil {
			return resultErr
		}
		r, err := resultsToInterfaces(resultVals)
		if err != nil {
			return err
		}
		r = append([]interface{}{prefix}, r...)
		_, err = fmt.Fprintln(writer, r...)
		return err
	}
}

func PrintlnWithPrefix(prefix string) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
		if resultErr != nil {
			return resultErr
		}
		r, err := resultsToInterfaces(resultVals)
		if err != nil {
			return err
		}
		r = append([]interface{}{prefix}, r...)
		_, err = fmt.Println(r...)
		return err
	}
}

type Logger interface {
	Printf(format string, args ...interface{})
}

func LogTo(logger Logger) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
		if resultErr != nil {
			return resultErr
		}
		r, err := resultsToInterfaces(resultVals)
		if err != nil || len(r) == 0 {
			return err
		}
		logger.Printf(fmt.Sprintln(r...))
		return nil
	}
}

func LogWithPrefixTo(prefix string, logger Logger) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
		if resultErr != nil {
			return resultErr
		}
		r, err := resultsToInterfaces(resultVals)
		if err != nil || len(r) == 0 {
			return err
		}
		r = append([]interface{}{prefix}, r...)
		logger.Printf(fmt.Sprintln(r...))
		return nil
	}
}

type PrintlnText string

func (t PrintlnText) HandleResults(args Args, argVals, resultVals []reflect.Value, resultErr error) error {
	if resultErr != nil {
		return resultErr
	}
	_, err := fmt.Println(t)
	return err
}
