package command

import (
	"reflect"
)

type Args interface {
	NumArgs() int
	Args() []Arg
	ArgTag(index int, tag string) string
	String() string
}

type Arg struct {
	Name        string
	Description string
	Type        reflect.Type
}
