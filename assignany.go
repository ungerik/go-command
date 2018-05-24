package command

import (
	"reflect"

	"github.com/pkg/errors"
	reflection "github.com/ungerik/go-reflection"
)

func assignAny(destVal reflect.Value, source interface{}) (err error) {
	if !destVal.CanSet() {
		return errors.Errorf("assignAny: destVal is not assignable")
	}

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = errors.Errorf("assignAny: %+v", r)
	// 	}
	// }()

	destType := destVal.Type()

	sourceVal := reflection.DerefValue(source)
	sourceType := sourceVal.Type()

	if sourceType == destType {
		destVal.Set(sourceVal)
		return nil
	}

	// TODO
	panic("NOT IMPLEMENTED")

	// switch destType.Kind() {
	// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	// 	if destType.Size() < sourceType.Size() {
	// 		return
	// 	}
	// 	destVal.SetInt(sourceVal.Int())

	// case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	// 	destVal.SetUint(sourceVal.Uint())

	// case reflect.Float32, reflect.Float64:
	// 	destVal.SetFloat(sourceVal.Float())

	// case reflect.String:
	// 	destVal.SetString(sourceVal.String())

	// case reflect.Bool:
	// 	destVal.SetBool(sourceVal.Bool())

	// default:
	// }

	// return nil
}
