package command

import (
	"context"
)

type StringArgsFunc func(ctx context.Context, args ...string) error

func NewStringArgsFunc(f Function, resultsHandlers ...ResultsHandler) StringArgsFunc {
	return func(ctx context.Context, args ...string) error {
		results, resultErr := f.CallWithStrings(ctx, args...)
		for _, resultsHandler := range resultsHandlers {
			err := resultsHandler.HandleResults(results, resultErr)
			if err != nil && err != resultErr {
				return err
			}
		}
		return resultErr
	}
}

type NamedStringArgsFunc func(ctx context.Context, args map[string]string) error

func NewNamedStringArgsFunc(f Function, resultsHandlers ...ResultsHandler) NamedStringArgsFunc {
	return func(ctx context.Context, args map[string]string) error {
		results, resultErr := f.CallWithNamedStrings(ctx, args)
		for _, resultsHandler := range resultsHandlers {
			err := resultsHandler.HandleResults(results, resultErr)
			if err != nil && err != resultErr {
				return err
			}
		}
		return resultErr
	}
}

type JSONArgsFunc func(ctx context.Context, jsonObject []byte) error

func NewJSONArgsFunc(f Function, resultsHandlers ...ResultsHandler) JSONArgsFunc {
	return func(ctx context.Context, jsonObject []byte) error {
		results, resultErr := CallFunctionWithJSONArgs(ctx, f, jsonObject)
		for _, resultsHandler := range resultsHandlers {
			err := resultsHandler.HandleResults(results, resultErr)
			if err != nil && err != resultErr {
				return err
			}
		}
		return resultErr
	}
}
