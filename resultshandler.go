package command

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/ungerik/go-reflection"
)

type ResultsHandler interface {
	HandleResults(args Args, argVals, resultVals []reflect.Value) error
}

type ResultsHandlerFunc func(args Args, argVals, resultVals []reflect.Value) error

func (f ResultsHandlerFunc) HandleResults(args Args, argVals, resultVals []reflect.Value) error {
	return f(args, argVals, resultVals)
}

func resultsToInterfaces(results []reflect.Value) ([]interface{}, error) {
	r := make([]interface{}, len(results))
	for i, result := range results {
		switch reflection.DerefValue(result).Kind() {
		case reflect.Struct, reflect.Slice, reflect.Array:
			b, err := json.MarshalIndent(result.Interface(), "", "  ")
			if err != nil {
				return nil, err
			}
			r[i] = string(b)
		default:
			r[i] = result.Interface()
		}
	}
	return r, nil
}

func PrintTo(writer io.Writer) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value) error {
		r, err := resultsToInterfaces(resultVals)
		if err != nil || len(r) == 0 {
			return err
		}
		_, err = fmt.Fprint(writer, r...)
		return err
	}
}

func PrintlnTo(writer io.Writer) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value) error {
		r, err := resultsToInterfaces(resultVals)
		if err != nil || len(r) == 0 {
			return err
		}
		_, err = fmt.Fprintln(writer, r...)
		return err
	}
}

func LogTo(logger *log.Logger) ResultsHandlerFunc {
	return func(args Args, argVals, resultVals []reflect.Value) error {
		r, err := resultsToInterfaces(resultVals)
		if err != nil || len(r) == 0 {
			return err
		}
		logger.Println(r...)
		return nil
	}
}

type PrintlnText string

func (t PrintlnText) HandleResults(args Args, argVals, resultVals []reflect.Value) error {
	_, err := fmt.Println(t)
	return err
}
