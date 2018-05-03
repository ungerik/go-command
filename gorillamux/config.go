package gorillamux

import "github.com/ungerik/go-httpx/httperr"

var (
	CatchPanics         = true
	PrettyPrint         = true
	PrettyPrintIndent   = "  "
	DefaultErrorHandler = httperr.HandlerFunc(httperr.DefaultHandlerFunc)
)
