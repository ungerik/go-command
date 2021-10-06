package gorillamux

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/ungerik/go-command"
	"github.com/ungerik/go-httpx/httperr"
)

func CommandHandler(commandFunc command.Function, resultsWriter ResultsWriter, errHandlers ...httperr.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if CatchPanics {
			defer func() {
				handleErr(httperr.AsError(recover()), writer, request, errHandlers)
			}()
		}

		vars := mux.Vars(request)

		results, err := commandFunc.CallWithNamedStrings(request.Context(), vars)

		if resultsWriter != nil {
			err = resultsWriter.WriteResults(results, err, writer, request)
		}
		handleErr(err, writer, request, errHandlers)
	}
}

func CommandHandlerWithQueryParams(commandFunc command.Function, resultsWriter ResultsWriter, errHandlers ...httperr.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if CatchPanics {
			defer func() {
				handleErr(httperr.AsError(recover()), writer, request, errHandlers)
			}()
		}

		vars := mux.Vars(request)

		// Add query params as arguments by joining them together per key
		// (query param names are not unique).
		for k := range request.URL.Query() {
			if len(request.URL.Query()[k]) > 0 && len(request.URL.Query()[k][0]) > 0 {
				vars[k] = strings.Join(request.URL.Query()[k][:], ";")
			}
		}

		results, err := commandFunc.CallWithNamedStrings(request.Context(), vars)

		if resultsWriter != nil {
			err = resultsWriter.WriteResults(results, err, writer, request)
		}
		handleErr(err, writer, request, errHandlers)
	}
}

type RequestBodyArgConverter interface {
	RequestBodyToArg(request *http.Request) (name, value string, err error)
}

type RequestBodyArgConverterFunc func(request *http.Request) (name, value string, err error)

func (f RequestBodyArgConverterFunc) RequestBodyToArg(request *http.Request) (name, value string, err error) {
	return f(request)
}

func RequestBodyAsArg(name string) RequestBodyArgConverterFunc {
	return func(request *http.Request) (string, string, error) {
		defer request.Body.Close()
		b, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return "", "", err
		}
		return name, string(b), nil
	}
}

func CommandHandlerRequestBodyArg(bodyConverter RequestBodyArgConverter, commandFunc command.Function, resultsWriter ResultsWriter, errHandlers ...httperr.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if CatchPanics {
			defer func() {
				handleErr(httperr.AsError(recover()), writer, request, errHandlers)
			}()
		}

		vars := mux.Vars(request)
		name, value, err := bodyConverter.RequestBodyToArg(request)
		if err != nil {
			handleErr(err, writer, request, errHandlers)
			return
		}
		if _, exists := vars[name]; exists {
			err = fmt.Errorf("argument '%s' already set by request URL path", name)
			handleErr(err, writer, request, errHandlers)
			return
		}
		vars[name] = value

		results, err := commandFunc.CallWithNamedStrings(request.Context(), vars)

		if resultsWriter != nil {
			err = resultsWriter.WriteResults(results, err, writer, request)
		}
		handleErr(err, writer, request, errHandlers)
	}
}

func handleErr(err error, writer http.ResponseWriter, request *http.Request, errHandlers []httperr.Handler) {
	if err == nil {
		return
	}
	if len(errHandlers) == 0 {
		httperr.Handle(err, writer, request)
	} else {
		for _, errHandler := range errHandlers {
			errHandler.HandleError(err, writer, request)
		}
	}
}

func MapJSONBodyFieldsAsVars(mapping map[string]string, wrappedHandler http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			httperr.BadRequest.ServeHTTP(writer, request)
			return
		}
		vars := mux.Vars(request)
		err = jsonBodyFieldsAsVars(body, mapping, vars)
		if err != nil {
			httperr.BadRequest.ServeHTTP(writer, request)
			return
		}
		wrappedHandler.ServeHTTP(writer, request)
	}
}

func JSONBodyFieldsAsVars(wrappedHandler http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			httperr.BadRequest.ServeHTTP(writer, request)
			return
		}
		vars := mux.Vars(request)
		err = jsonBodyFieldsAsVars(body, nil, vars)
		if err != nil {
			httperr.BadRequest.ServeHTTP(writer, request)
			return
		}
		wrappedHandler.ServeHTTP(writer, request)
	}
}

func jsonBodyFieldsAsVars(body []byte, mapping map[string]string, vars map[string]string) error {
	fields := make(map[string]json.RawMessage)
	err := json.Unmarshal(body, &fields)
	if err != nil {
		return err
	}

	if mapping != nil {
		mappedFields := make(map[string]json.RawMessage, len(fields))
		for fieldName, mappedName := range mapping {
			if value, ok := fields[fieldName]; ok {
				mappedFields[mappedName] = value
			}
		}
		fields = mappedFields
	}

	for name, value := range fields {
		if len(value) == 0 {
			// should never happen with well formed JSON
			return fmt.Errorf("JSON body field %q is empty", name)
		}
		valueStr := string(value)
		switch {
		case valueStr == "null":
			// JSON nulls are left alone

		case valueStr[0] == '"':
			// Unescape JSON string
			err = json.Unmarshal(value, &valueStr)
			if err != nil {
				return fmt.Errorf("can't unmarshal JSON body field %q as string because of: %w", name, err)
			}
			vars[name] = valueStr

		default:
			// All other JSON types are mapped directly to string
			vars[name] = valueStr
		}
	}
	return nil
}
