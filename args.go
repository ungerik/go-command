package command

import (
	"reflect"
)

type Args interface {
	NumArgs() int
	ArgName(index int) string
	ArgDescription(index int) string
	ArgTag(index int, tag string) string
	ArgType(index int) reflect.Type
	String() string
}
