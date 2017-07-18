package gorillamux

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"

	"github.com/ungerik/go-command"
	"github.com/ungerik/go-httpx/returning"
)

type ResultsWriter interface {
	WriteResults(results []reflect.Value, writer http.ResponseWriter, request *http.Request) error
	WriteError(err error, writer http.ResponseWriter, request *http.Request)
}

var RespondJSON respondJSON

type respondJSON struct{}

func (respondJSON) WriteResults(results []reflect.Value, writer http.ResponseWriter, request *http.Request) error {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	for _, result := range results {
		encoder := json.NewEncoder(buf)
		if returning.PrettyPrintResponses {
			encoder.SetIndent("", returning.PrettyPrintIndent)
		}
		err := encoder.Encode(result)
		if err != nil {
			return err
		}
	}
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(buf.Bytes())
	return nil
}

func (respondJSON) WriteError(err error, writer http.ResponseWriter, request *http.Request) {
	returning.HandleError(err, writer, request)
}

func resultsHandlerFrom(resultsWriter ResultsWriter, writer http.ResponseWriter, request *http.Request) command.ResultsHandlerFunc {
	return func(results []reflect.Value) error {
		return resultsWriter.WriteResults(results, writer, request)
	}
}

func CommandHandler(args command.Args, commandFunc interface{}, resultsWriter ResultsWriter) http.HandlerFunc {
	f := command.MustGetStringMapArgsWithResultsHandlerFunc(args, commandFunc)
	return func(writer http.ResponseWriter, request *http.Request) {
		if returning.CatchPanics {
			defer func() {
				if r := recover(); r != nil {
					returning.WriteInternalServerError(writer, r)
				}
			}()
		}
		args := mux.Vars(request)
		resultsHandler := resultsHandlerFrom(resultsWriter, writer, request)
		err := f(args, resultsHandler)
		if err != nil {
			resultsWriter.WriteError(err, writer, request)
		}
	}
}
