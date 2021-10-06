package command

import "fmt"

type ErrParseArgString struct {
	Err  error
	Func FunctionInfo
	Arg  string
}

func NewErrParseArgString(err error, f FunctionInfo, arg string) ErrParseArgString {
	return ErrParseArgString{Err: err, Func: f, Arg: arg}
}

func (e ErrParseArgString) Error() string {
	return fmt.Sprintf("string conversion error for argument %s of function %s: %s", e.Arg, e.Func, e.Err)
}

func (e ErrParseArgString) Unwrap() error {
	return e.Err
}

type ErrParseArgJSON struct {
	Err  error
	Func FunctionInfo
	Arg  string
}

func NewErrParseArgJSON(err error, f FunctionInfo, arg string) ErrParseArgJSON {
	return ErrParseArgJSON{Err: err, Func: f, Arg: arg}
}

func (e ErrParseArgJSON) Error() string {
	return fmt.Sprintf("error unmarshalling JSON for argument %s of function %s: %s", e.Arg, e.Func, e.Err)
}

func (e ErrParseArgJSON) Unwrap() error {
	return e.Err
}
