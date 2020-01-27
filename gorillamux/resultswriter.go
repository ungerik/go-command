package gorillamux

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"

	command "github.com/ungerik/go-command"
)

type ResultsWriter interface {
	WriteResults(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) error
}

type ResultsWriterFunc func(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) error

func (f ResultsWriterFunc) WriteResults(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	return f(args, vars, resultVals, resultErr, writer, request)
}

func encodeJSON(response interface{}) ([]byte, error) {
	if PrettyPrint {
		return json.MarshalIndent(response, "", PrettyPrintIndent)
	}
	return json.Marshal(response)
}

var RespondJSON ResultsWriterFunc = func(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	if resultErr != nil {
		return resultErr
	}
	var buf []byte
	for _, resultVal := range resultVals {
		b, err := encodeJSON(resultVal.Interface())
		if err != nil {
			return err
		}
		buf = append(buf, b...)
	}
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(buf)
	return nil
}

// RespondBinary responds with contentType using the binary data from results of type []byte, string, or io.Reader.
func RespondBinary(contentType string) ResultsWriterFunc {
	return func(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) (err error) {
		if resultErr != nil {
			return resultErr
		}
		var buf bytes.Buffer
		for _, resultVal := range resultVals {
			switch data := resultVal.Interface().(type) {
			case []byte:
				_, err = buf.Write(data)
			case string:
				_, err = buf.WriteString(data)
			case io.Reader:
				_, err = io.Copy(&buf, data)
			default:
				return fmt.Errorf("RespondBinary does not support result type %s", resultVal.Type())
			}
			if err != nil {
				return err
			}
		}
		writer.Header().Set("Content-Type", contentType)
		writer.Write(buf.Bytes())
		return nil
	}
}

func RespondJSONField(fieldName string) ResultsWriterFunc {
	return func(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) (err error) {
		if resultErr != nil {
			return resultErr
		}
		var buf []byte
		m := make(map[string]interface{})
		if len(resultVals) > 0 {
			m[fieldName] = resultVals[0].Interface()
		}
		buf, err = encodeJSON(m)
		if err != nil {
			return err
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.Write(buf)
		return nil
	}
}

func encodeXML(response interface{}) ([]byte, error) {
	if PrettyPrint {
		return xml.MarshalIndent(response, "", PrettyPrintIndent)
	}
	return xml.Marshal(response)
}

var RespondXML ResultsWriterFunc = func(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	if resultErr != nil {
		return resultErr
	}
	var buf []byte
	for _, resultVal := range resultVals {
		b, err := encodeXML(resultVal.Interface())
		if err != nil {
			return err
		}
		buf = append(buf, b...)
	}
	writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	writer.Write(buf)
	return nil
}

var RespondPlaintext ResultsWriterFunc = func(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	if resultErr != nil {
		return resultErr
	}
	var buf bytes.Buffer
	for _, resultVal := range resultVals {
		result := resultVal.Interface()
		if b, ok := result.([]byte); ok {
			buf.Write(b)
		} else {
			fmt.Fprint(&buf, result)
		}
	}
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.Write(buf.Bytes())
	return nil
}

var RespondHTML ResultsWriterFunc = func(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	if resultErr != nil {
		return resultErr
	}
	var buf bytes.Buffer
	for _, resultVal := range resultVals {
		result := resultVal.Interface()
		if b, ok := result.([]byte); ok {
			buf.Write(b)
		} else {
			fmt.Fprint(&buf, result)
		}
	}
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.Write(buf.Bytes())
	return nil
}

var RespondDetectContentType ResultsWriterFunc = func(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	if resultErr != nil {
		return resultErr
	}
	if len(resultVals) != 1 {
		return fmt.Errorf("RespondDetectContentType needs 1 result, got %d", len(resultVals))
	}
	data, ok := resultVals[0].Interface().([]byte)
	if !ok {
		return fmt.Errorf("RespondDetectContentType needs []byte result, got %s", resultVals[0].Type())
	}

	writer.Header().Add("Content-Type", DetectContentType(data))
	writer.Write(data)
	return nil
}

func RespondContentType(contentType string) ResultsWriter {
	return ResultsWriterFunc(func(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) error {
		if resultErr != nil {
			return resultErr
		}
		if len(resultVals) != 1 {
			return fmt.Errorf("RespondDetectContentType needs 1 result, got %d", len(resultVals))
		}
		data, ok := resultVals[0].Interface().([]byte)
		if !ok {
			return fmt.Errorf("RespondDetectContentType needs []byte result, got %s", resultVals[0].Type())
		}

		writer.Header().Add("Content-Type", contentType)
		writer.Write(data)
		return nil
	})
}

var RespondNothing ResultsWriterFunc = func(args command.Args, vars map[string]string, resultVals []reflect.Value, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	return resultErr
}

// DetectContentType tries to detect the MIME content-type of data,
// or returns "application/octet-stream" if none could be identified.
func DetectContentType(data []byte) string {
	kind, _ := filetype.Match(data)
	if kind == types.Unknown {
		return http.DetectContentType(data)
	}
	return kind.MIME.Value
}
