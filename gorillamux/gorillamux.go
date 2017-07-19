package gorillamux

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ungerik/go-command"
	"github.com/ungerik/go-httpx/httperr"
)

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

func CommandHandler(args command.Args, commandFunc interface{}, resultsWriter ResultsWriter, errHandlers ...httperr.Handler) http.HandlerFunc {
	f := command.MustGetStringMapArgsResultValuesFunc(args, commandFunc)

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
