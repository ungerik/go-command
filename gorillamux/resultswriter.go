package gorillamux

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
)

type ResultsWriter interface {
	WriteResults(results []reflect.Value, writer http.ResponseWriter, request *http.Request) error
}

type ResultsWriterFunc func(results []reflect.Value, writer http.ResponseWriter, request *http.Request) error

func (f ResultsWriterFunc) WriteResults(results []reflect.Value, writer http.ResponseWriter, request *http.Request) error {
	return f(results, writer, request)
}

func encodeJSON(response interface{}) ([]byte, error) {
	if PrettyPrint {
		return json.MarshalIndent(response, "", PrettyPrintIndent)
	}
	return json.Marshal(response)
}

var RespondJSON ResultsWriterFunc = func(results []reflect.Value, writer http.ResponseWriter, request *http.Request) error {
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

var RespondXML ResultsWriterFunc = func(results []reflect.Value, writer http.ResponseWriter, request *http.Request) error {
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

var RespondPlaintext ResultsWriterFunc = func(results []reflect.Value, writer http.ResponseWriter, request *http.Request) error {
	var buf bytes.Buffer
	for _, result := range results {
		fmt.Fprintf(&buf, "%s", result.Interface())
	}
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.Write(buf.Bytes())
	return nil
}

var RespondHTML ResultsWriterFunc = func(results []reflect.Value, writer http.ResponseWriter, request *http.Request) error {
	var buf bytes.Buffer
	for _, result := range results {
		fmt.Fprintf(&buf, "%s", result.Interface())
	}
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.Write(buf.Bytes())
	return nil
}
