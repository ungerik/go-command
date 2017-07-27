package command

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

func assignString(destVal reflect.Value, sourceStr string) error {
	destPtr := destVal.Addr().Interface()

	switch v := destPtr.(type) {
	case *string:
		*v = sourceStr
		return nil

	case *[]byte:
		*v = []byte(sourceStr)
		return nil

	case encoding.TextUnmarshaler:
		return v.UnmarshalText([]byte(sourceStr))
	}

	switch destVal.Kind() {
	case reflect.String:
		destVal.Set(reflect.ValueOf(sourceStr).Convert(destVal.Type()))
		return nil

	case reflect.Struct:
		// JSON might not be the best format for command line arguments,
		// but it could have also come from a HTTP request body or other sources
		return json.Unmarshal([]byte(sourceStr), destPtr)

	case reflect.Ptr:
		ptr := destVal
		if ptr.IsNil() {
			ptr = reflect.New(destVal.Type().Elem())
		}
		err := assignString(ptr.Elem(), sourceStr)
		if err != nil {
			return err
		}
		destVal.Set(ptr)
		return nil

	case reflect.Slice:
		if !strings.HasPrefix(sourceStr, "[") {
			return errors.Errorf("Slice value '%s' does not begin with '['", sourceStr)
		}
		if !strings.HasSuffix(sourceStr, "]") {
			return errors.Errorf("Slice value '%s' does not end with ']'", sourceStr)
		}
		// elemSourceStrings := strings.Split(sourceStr[1:len(sourceStr)-1], ",")
		sourceFields, err := sliceLiteralFields(sourceStr)
		if err != nil {
			return err
		}

		count := len(sourceFields)
		destVal.Set(reflect.MakeSlice(destVal.Type(), count, count))

		for i := 0; i < count; i++ {
			err := assignString(destVal.Index(i), sourceFields[i])
			if err != nil {
				return err
			}
		}
		return nil

	case reflect.Array:
		if !strings.HasPrefix(sourceStr, "[") {
			return errors.Errorf("Array value '%s' does not begin with '['", sourceStr)
		}
		if !strings.HasSuffix(sourceStr, "]") {
			return errors.Errorf("Array value '%s' does not end with ']'", sourceStr)
		}
		// elemSourceStrings := strings.Split(sourceStr[1:len(sourceStr)-1], ",")
		sourceFields, err := sliceLiteralFields(sourceStr)
		if err != nil {
			return err
		}

		count := len(sourceFields)
		if count != destVal.Len() {
			return errors.Errorf("Array value '%s' needs to have %d elements, but has %d", sourceStr, destVal.Len(), count)

		}

		for i := 0; i < count; i++ {
			err := assignString(destVal.Index(i), sourceFields[i])
			if err != nil {
				return err
			}
		}
		return nil
	}

	// If all else fails, use fmt scanning
	// for generic type conversation from string
	_, err := fmt.Sscan(sourceStr, destPtr)
	return err
}

func sliceLiteralFields(sourceStr string) (fields []string, err error) {
	if !strings.HasPrefix(sourceStr, "[") {
		return nil, errors.Errorf("Slice value '%s' does not begin with '['", sourceStr)
	}
	if !strings.HasSuffix(sourceStr, "]") {
		return nil, errors.Errorf("Slice value '%s' does not end with ']'", sourceStr)
	}
	bracketDepth := 0
	begin := 1
	for i, r := range sourceStr {
		switch r {
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
			if bracketDepth < 0 {
				return nil, errors.Errorf("Slice value '%s' has too many ']'", sourceStr)
			}
			if bracketDepth == 0 && i-begin > 0 {
				fields = append(fields, sourceStr[begin:i])
			}
		case ',':
			if bracketDepth == 1 {
				fields = append(fields, sourceStr[begin:i])
				begin = i + 1
			}
		}
	}
	return fields, nil
}