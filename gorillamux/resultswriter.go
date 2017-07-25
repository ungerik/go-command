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
	WriteResults(args command.Args, vars map[string]string, results []reflect.Value, writer http.ResponseWriter, request *http.Request) error

	// HandleError can handle err and return nil,
	// or return err if it does not want to handle it.
	HandleError(err error) error
}

type ResultsWriterFunc func(args command.Args, vars map[string]string, results []reflect.Value, writer http.ResponseWriter, request *http.Request) error

func (f ResultsWriterFunc) WriteResults(args command.Args, vars map[string]string, results []reflect.Value, writer http.ResponseWriter, request *http.Request) error {
	return f(args, vars, results, writer, request)
}

func (f ResultsWriterFunc) HandleError(err error) error {
	return err
}

func encodeJSON(response interface{}) ([]byte, error) {
	if PrettyPrint {
		return json.MarshalIndent(response, "", PrettyPrintIndent)
	}
	return json.Marshal(response)
}

var RespondJSON ResultsWriterFunc = func(args command.Args, vars map[string]string, results []reflect.Value, writer http.ResponseWriter, request *http.Request) error {
	var buf []byte
	for _, result := range results {
		b, err := encodeJSON(result.Interface())
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

var RespondXML ResultsWriterFunc = func(args command.Args, vars map[string]string, results []reflect.Value, writer http.ResponseWriter, request *http.Request) error {
	var buf []byte
	for _, result := range results {
		b, err := encodeXML(result.Interface())
		if err != nil {
			return err
		}
		buf = append(buf, b...)
	}
	writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	writer.Write(buf)
	return nil
}

var RespondPlaintext ResultsWriterFunc = func(args command.Args, vars map[string]string, results []reflect.Value, writer http.ResponseWriter, request *http.Request) error {
	var buf bytes.Buffer
	for _, result := range results {
		fmt.Fprintf(&buf, "%s", result.Interface())
	}
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.Write(buf.Bytes())
	return nil
}

var RespondHTML ResultsWriterFunc = func(args command.Args, vars map[string]string, results []reflect.Value, writer http.ResponseWriter, request *http.Request) error {
	var buf bytes.Buffer
	for _, result := range results {
		fmt.Fprintf(&buf, "%s", result.Interface())
	}
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.Write(buf.Bytes())
	return nil
}
