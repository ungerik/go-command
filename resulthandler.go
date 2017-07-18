package command

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/ungerik/go-reflection"
)

func PrintTo(writer io.Writer) ResultHandlerFunc {
	return func(result reflect.Value) error {
		switch reflection.DerefValue(result).Kind() {
		case reflect.Struct, reflect.Slice, reflect.Array:
			enc := json.NewEncoder(writer)
			enc.SetIndent("", "  ")
			return enc.Encode(result.Interface())
		}
		_, err := fmt.Fprintln(writer, result.Interface())
		return err
	}
}

func LogTo(logger *log.Logger) ResultHandlerFunc {
	return func(result reflect.Value) error {
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
		return nil
	}
}
