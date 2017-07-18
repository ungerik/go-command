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
	HandleResults(results []reflect.Value) error
}

type ResultsHandlerFunc func(results []reflect.Value) error

func (f ResultsHandlerFunc) HandleResults(results []reflect.Value) error {
	return f(results)
}

// TODO single print call?
func PrintTo(writer io.Writer) ResultsHandlerFunc {
	return func(results []reflect.Value) (err error) {
		for _, result := range results {
			switch reflection.DerefValue(result).Kind() {
			case reflect.Struct, reflect.Slice, reflect.Array:
				b, err := json.MarshalIndent(result.Interface(), "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprint(writer, string(b))
			default:
				fmt.Fprint(writer, result.Interface())
			}
		}
		return nil
	}
}

func PrintlnTo(writer io.Writer) ResultsHandlerFunc {
	return func(results []reflect.Value) (err error) {
		for _, result := range results {
			switch reflection.DerefValue(result).Kind() {
			case reflect.Struct, reflect.Slice, reflect.Array:
				b, err := json.MarshalIndent(result.Interface(), "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(writer, string(b))
			default:
				fmt.Fprintln(writer, result.Interface())
			}
		}
		return nil
	}
}

func LogTo(logger *log.Logger) ResultsHandlerFunc {
	return func(results []reflect.Value) error {
		for _, result := range results {
			switch reflection.DerefValue(result).Kind() {
			case reflect.Struct, reflect.Slice, reflect.Array:
				b, err := json.Marshal(result.Interface())
				if err != nil {
					return err
				}
				logger.Print(string(b))
			default:
				logger.Print(result.Interface())
			}
		}
		return nil
	}
}
