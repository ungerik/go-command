package command

import "fmt"

type ErrArgFromString struct {
	Err  error
	Func Function
	Arg  string
}

func NewErrArgFromString(err error, f Function, arg string) ErrArgFromString {
	return ErrArgFromString{Err: err, Func: f, Arg: arg}
}

func (e ErrArgFromString) Error() string {
	return fmt.Sprintf("string conversion error for argument %s of function %s: %s", e.Arg, e.Func, e.Err)
}

func (e ErrArgFromString) Unwrap() error {
	return e.Err
}
