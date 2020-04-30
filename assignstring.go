package command

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	fs "github.com/ungerik/go-fs"
)

func assignString(destVal reflect.Value, sourceStr string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("assignString(%s, %q): %w", destVal.Type(), sourceStr, err)
		}
	}()

	destPtr := destVal.Addr().Interface()

	switch dest := destPtr.(type) {
	case *string:
		*dest = sourceStr
		return nil

	case *time.Time:
		for _, format := range TimeFormats {
			t, err := time.Parse(format, sourceStr)
			if err == nil {
				*dest = t
				return nil
			}
		}
		return fmt.Errorf("can't parse %q as time.Time using formats %#v", sourceStr, TimeFormats)

	case *time.Duration:
		duration, err := time.ParseDuration(sourceStr)
		if err != nil {
			return fmt.Errorf("can't parse %q as time.Duration because of: %w", sourceStr, err)
		}
		*dest = duration
		return nil

	case encoding.TextUnmarshaler:
		return dest.UnmarshalText([]byte(sourceStr))

	case *fs.FileReader:
		*dest = fs.File(sourceStr)
		return nil

	case json.Unmarshaler:
		return dest.UnmarshalJSON([]byte(sourceStr))

	case *map[string]interface{}:
		return json.Unmarshal([]byte(sourceStr), dest)

	case *[]interface{}:
		return json.Unmarshal([]byte(sourceStr), dest)

	case *[]byte:
		*dest = []byte(sourceStr)
		return nil
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
		if sourceStr != "nil" {
			if ptr.IsNil() {
				ptr = reflect.New(destVal.Type().Elem())
			}
			err := assignString(ptr.Elem(), sourceStr)
			if err != nil {
				return err
			}
			destVal.Set(ptr)
		}
		return nil

	case reflect.Slice:
		if !strings.HasPrefix(sourceStr, "[") {
			return fmt.Errorf("slice value %q does not begin with '['", sourceStr)
		}
		if !strings.HasSuffix(sourceStr, "]") {
			return fmt.Errorf("slice value %q does not end with ']'", sourceStr)
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
			return fmt.Errorf("array value %q does not begin with '['", sourceStr)
		}
		if !strings.HasSuffix(sourceStr, "]") {
			return fmt.Errorf("array value %q does not end with ']'", sourceStr)
		}
		// elemSourceStrings := strings.Split(sourceStr[1:len(sourceStr)-1], ",")
		sourceFields, err := sliceLiteralFields(sourceStr)
		if err != nil {
			return err
		}

		count := len(sourceFields)
		if count != destVal.Len() {
			return fmt.Errorf("array value %q needs to have %d elements, but has %d", sourceStr, destVal.Len(), count)

		}

		for i := 0; i < count; i++ {
			err := assignString(destVal.Index(i), sourceFields[i])
			if err != nil {
				return err
			}
		}
		return nil

	case reflect.Func:
		// We can't assign a string to a function, it's OK to ignore it
		return nil
	}

	// If all else fails, use fmt scanning
	// for generic type conversation from string
	_, err = fmt.Sscan(sourceStr, destPtr)
	return err
}

func sliceLiteralFields(sourceStr string) (fields []string, err error) {
	if !strings.HasPrefix(sourceStr, "[") {
		return nil, fmt.Errorf("slice value %q does not begin with '['", sourceStr)
	}
	if !strings.HasSuffix(sourceStr, "]") {
		return nil, fmt.Errorf("slice value %q does not end with ']'", sourceStr)
	}
	objectDepth := 0
	bracketDepth := 0
	begin := 1
	for i, r := range sourceStr {
		switch r {
		case '{':
			objectDepth++

		case '}':
			objectDepth--
			if objectDepth < 0 {
				return nil, fmt.Errorf("slice value %q has too many '}'", sourceStr)
			}

		case '[':
			bracketDepth++

		case ']':
			bracketDepth--
			if bracketDepth < 0 {
				return nil, fmt.Errorf("slice value %q has too many ']'", sourceStr)
			}
			if objectDepth == 0 && bracketDepth == 0 && i-begin > 0 {
				fields = append(fields, sourceStr[begin:i])
			}

		case ',':
			if objectDepth == 0 && bracketDepth == 1 {
				fields = append(fields, sourceStr[begin:i])
				begin = i + 1
			}
		}
	}
	return fields, nil
}
