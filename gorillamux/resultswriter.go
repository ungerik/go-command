package gorillamux

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"

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
		fmt.Fprintf(&buf, "%s", resultVal.Interface())
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
		fmt.Fprintf(&buf, "%s", resultVal.Interface())
	}
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.Write(buf.Bytes())
	return nil
}
