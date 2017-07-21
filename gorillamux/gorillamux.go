package gorillamux

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/ungerik/go-command"
	"github.com/ungerik/go-httpx/httperr"
)

func CommandHandler(commandFunc interface{}, args command.Args, resultsWriter ResultsWriter, errHandlers ...httperr.Handler) http.HandlerFunc {
	f := command.MustGetStringMapArgsResultValuesFunc(commandFunc, args)

	return func(writer http.ResponseWriter, request *http.Request) {
		if CatchPanics {
			defer handleErr(httperr.Recover(), writer, request, errHandlers)
		}

		results, err := f(mux.Vars(request))

		if err != nil {
			handleErr(err, writer, request, errHandlers)
		} else {
			resultsWriter.WriteResults(results, writer, request)
		}
	}
}

type RequestBodyArgConverter interface {
	RequestBodyToArg(request *http.Request) (name, value string, err error)
}

type RequestBodyArgConverterFunc func(request *http.Request) (name, value string, err error)

func (f RequestBodyArgConverterFunc) RequestBodyToArg(request *http.Request) (name, value string, err error) {
	return f(request)
}

func JSONRequestBodyAsArg(name string) RequestBodyArgConverterFunc {
	return func(request *http.Request) (name, value string, err error) {
		defer request.Body.Close()
		b, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return "", "", err
		}
		return name, string(b), nil
	}
}

func CommandHandlerRequestBodyArg(bodyConverter RequestBodyArgConverter, commandFunc interface{}, args command.Args, resultsWriter ResultsWriter, errHandlers ...httperr.Handler) http.HandlerFunc {
	f := command.MustGetStringMapArgsResultValuesFunc(commandFunc, args)

	return func(writer http.ResponseWriter, request *http.Request) {
		if CatchPanics {
			defer handleErr(httperr.Recover(), writer, request, errHandlers)
		}

		args := mux.Vars(request)
		name, value, err := bodyConverter.RequestBodyToArg(request)
		if err != nil {
			handleErr(err, writer, request, errHandlers)
			return
		}
		if _, exists := args[name]; exists {
			err = errors.Errorf("Argument '%s' already set by request URL path", name)
			handleErr(err, writer, request, errHandlers)
			return
		}
		args[name] = value

		results, err := f(args)

		if err != nil {
			handleErr(err, writer, request, errHandlers)
		} else {
			resultsWriter.WriteResults(results, writer, request)
		}
	}
}

func handleErr(err error, writer http.ResponseWriter, request *http.Request, errHandlers []httperr.Handler) {
	if err == nil {
		return
	}
	if len(errHandlers) == 0 {
		DefaultErrorHandler.HandleError(err, writer, request)
	} else {
		for _, errHandler := range errHandlers {
			errHandler.HandleError(err, writer, request)
		}
	}
}
