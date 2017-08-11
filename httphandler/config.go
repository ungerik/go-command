package httphandler

import "github.com/ungerik/go-httpx/httperr"

var (
	CatchPanics         bool
	PrettyPrint         bool
	PrettyPrintIndent   = "  "
	DefaultErrorHandler = httperr.HandlerFunc(httperr.DefaultHandlerFunc)
)
