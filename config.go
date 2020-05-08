package command

import (
	"context"
	"reflect"
	"time"

	"github.com/fatih/color"
)

var (
	// CommandUsageColor is the color in which the
	// command usage will be printed on the screen.
	CommandUsageColor = color.New(color.FgHiCyan)

	// CommandDescriptionColor is the color in which the
	// command usage description will be printed on the screen.
	CommandDescriptionColor = color.New(color.FgCyan)

	ArgNameTag        = "arg"
	ArgDescriptionTag = "desc"

	// TimeFormats used in that order to try parse time strings.
	// If a time format has not time zone part,
	// then the date is returned in the local time zone.
	TimeFormats = []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
)

var (
	typeOfError          = reflect.TypeOf((*error)(nil)).Elem()
	typeOfContext        = reflect.TypeOf((*context.Context)(nil)).Elem()
	typeOfEmptyInterface = reflect.TypeOf((*interface{})(nil)).Elem()
)
