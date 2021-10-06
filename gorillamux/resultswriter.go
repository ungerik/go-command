package gorillamux

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

type ResultsWriter interface {
	WriteResults(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) error
}

type ResultsWriterFunc func(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) error

func (f ResultsWriterFunc) WriteResults(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	return f(results, resultErr, writer, request)
}

func encodeJSON(response interface{}) ([]byte, error) {
	if PrettyPrint {
		return json.MarshalIndent(response, "", PrettyPrintIndent)
	}
	return json.Marshal(response)
}

var RespondJSON ResultsWriterFunc = func(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	if resultErr != nil {
		return resultErr
	}
	var buf []byte
	for _, result := range results {
		b, err := encodeJSON(result)
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
	return func(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) (err error) {
		if resultErr != nil {
			return resultErr
		}
		var buf bytes.Buffer
		for _, result := range results {
			switch data := result.(type) {
			case []byte:
				_, err = buf.Write(data)
			case string:
				_, err = buf.WriteString(data)
			case io.Reader:
				_, err = io.Copy(&buf, data)
			default:
				return fmt.Errorf("RespondBinary does not support result type %T", result)
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
	return func(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) (err error) {
		if resultErr != nil {
			return resultErr
		}
		var buf []byte
		m := make(map[string]interface{})
		if len(results) > 0 {
			m[fieldName] = results[0]
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

var RespondXML ResultsWriterFunc = func(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	if resultErr != nil {
		return resultErr
	}
	var buf []byte
	for _, result := range results {
		b, err := encodeXML(result)
		if err != nil {
			return err
		}
		buf = append(buf, b...)
	}
	writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	writer.Write(buf)
	return nil
}

var RespondPlaintext ResultsWriterFunc = func(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	if resultErr != nil {
		return resultErr
	}
	var buf bytes.Buffer
	for _, result := range results {
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

var RespondHTML ResultsWriterFunc = func(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	if resultErr != nil {
		return resultErr
	}
	var buf bytes.Buffer
	for _, result := range results {
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

var RespondDetectContentType ResultsWriterFunc = func(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) error {
	if resultErr != nil {
		return resultErr
	}
	if len(results) != 1 {
		return fmt.Errorf("RespondDetectContentType needs 1 result, got %d", len(results))
	}
	data, ok := results[0].([]byte)
	if !ok {
		return fmt.Errorf("RespondDetectContentType needs []byte result, got %T", results[0])
	}

	writer.Header().Add("Content-Type", DetectContentType(data))
	writer.Write(data)
	return nil
}

func RespondContentType(contentType string) ResultsWriter {
	return ResultsWriterFunc(func(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) error {
		if resultErr != nil {
			return resultErr
		}
		if len(results) != 1 {
			return fmt.Errorf("RespondDetectContentType needs 1 result, got %d", len(results))
		}
		data, ok := results[0].([]byte)
		if !ok {
			return fmt.Errorf("RespondDetectContentType needs []byte result, got %T", results[0])
		}

		writer.Header().Add("Content-Type", contentType)
		writer.Write(data)
		return nil
	})
}

var RespondNothing ResultsWriterFunc = func(results []interface{}, resultErr error, writer http.ResponseWriter, request *http.Request) error {
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
