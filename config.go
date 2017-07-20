package command

import "github.com/fatih/color"

var (
	// CommandUsageColor is the color in which the
	// command usage will be printed on the screen.
	CommandUsageColor = color.New(color.FgHiCyan)

	// CommandDescriptionColor is the color in which the
	// command usage description will be printed on the screen.
	CommandDescriptionColor = color.New(color.FgCyan)

	ArgNameTag        = "arg"
	ArgDescriptionTag = "desc"
)
