package gorillamux

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ungerik/go-command"
	"github.com/ungerik/go-httpx/returning"
)

func CommandHandler(commandFunc interface{}, args command.Args) returning.Error {
	stringMapArgsFunc := command.MustGetStringMapArgsFunc(args, commandFunc)
	return func(writer http.ResponseWriter, request *http.Request) error {
		return stringMapArgsFunc(mux.Vars(request))
	}
}

// func JSONCommandHandler(commandFunc interface{}, args command.Args) returning.JSON {
// 	stringMapArgsFunc := command.MustGetStringMapArgsFunc(args, commandFunc)
// 	return func(writer http.ResponseWriter, request *http.Request) error {
// 		return stringMapArgsFunc(mux.Vars(request))
// 	}
// }
