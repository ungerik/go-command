package command

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

type constError string

func (e constError) Error() string {
	return string(e)
}

const (
	Default = ""

	ErrNotFound = constError("command not found")
)

type stringArgsCommand struct {
	command         string
	description     string
	args            Args
	commandFunc     interface{}
	stringArgsFunc  StringArgsFunc
	resultsHandlers []ResultsHandler
}

func checkCommandChars(command string) error {
	if strings.IndexFunc(command, unicode.IsSpace) >= 0 {
		return errors.Errorf("Command contains space characters: '%s'", command)
	}
	if strings.IndexFunc(command, unicode.IsGraphic) == -1 {
		return errors.Errorf("Command contains non graphc characters: '%s'", command)
	}
	if strings.ContainsAny(command, "|&;()<>") {
		return errors.Errorf("Command contains invalid characters: '%s'", command)
	}
	return nil
}

type StringArgsCommandLogger interface {
	LogStringArgsCommand(command string, args []string)
}

type StringArgsCommandLoggerFunc func(command string, args []string)

func (f StringArgsCommandLoggerFunc) LogStringArgsCommand(command string, args []string) {
	f(command, args)
}

type StringArgsDispatcher struct {
	comm    map[string]*stringArgsCommand
	loggers []StringArgsCommandLogger
}

func NewStringArgsDispatcher(loggers ...StringArgsCommandLogger) *StringArgsDispatcher {
	return &StringArgsDispatcher{
		comm:    make(map[string]*stringArgsCommand),
		loggers: loggers,
	}
}

func (disp *StringArgsDispatcher) AddCommand(command, description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) error {
	if err := checkCommandChars(command); err != nil {
		return errors.Wrapf(err, "Command '%s'", command)
	}
	stringArgsFunc, err := GetStringArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		return errors.Wrapf(err, "Command '%s'", command)
	}
	disp.comm[command] = &stringArgsCommand{
		command:         command,
		description:     description,
		args:            args,
		commandFunc:     commandFunc,
		stringArgsFunc:  stringArgsFunc,
		resultsHandlers: resultsHandlers,
	}
	return nil
}

func (disp *StringArgsDispatcher) MustAddCommand(command, description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) {
	err := disp.AddCommand(command, description, commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
}

func (disp *StringArgsDispatcher) AddDefaultCommand(description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) error {
	stringArgsFunc, err := GetStringArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		return errors.Wrap(err, "Default command")
	}
	disp.comm[Default] = &stringArgsCommand{
		command:         Default,
		description:     description,
		args:            args,
		commandFunc:     commandFunc,
		stringArgsFunc:  stringArgsFunc,
		resultsHandlers: resultsHandlers,
	}
	return nil
}

func (disp *StringArgsDispatcher) MustAddDefaultCommand(description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) {
	err := disp.AddDefaultCommand(description, commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
}

func (disp *StringArgsDispatcher) HasCommnd(command string) bool {
	_, found := disp.comm[command]
	return found
}

func (disp *StringArgsDispatcher) HasDefaultCommnd() bool {
	_, found := disp.comm[Default]
	return found
}

func (disp *StringArgsDispatcher) Dispatch(command string, args ...string) error {
	cmd, found := disp.comm[command]
	if !found {
		return ErrNotFound
	}
	for _, logger := range disp.loggers {
		logger.LogStringArgsCommand(command, args)
	}
	return cmd.stringArgsFunc(args...)
}

func (disp *StringArgsDispatcher) MustDispatch(command string, args ...string) {
	err := disp.Dispatch(command, args...)
	if err != nil {
		panic(errors.Wrapf(err, "Command '%s'", command))
	}
}

func (disp *StringArgsDispatcher) DispatchDefaultCommand() error {
	return disp.Dispatch(Default)
}

func (disp *StringArgsDispatcher) MustDispatchDefaultCommand() {
	err := disp.DispatchDefaultCommand()
	if err != nil {
		panic(errors.Wrap(err, "Default command"))
	}
}

func (disp *StringArgsDispatcher) DispatchCombinedCommandAndArgs(commandAndArgs []string) (command string, err error) {
	if len(commandAndArgs) == 0 {
		return Default, disp.DispatchDefaultCommand()
	}
	command = commandAndArgs[0]
	args := commandAndArgs[1:]
	return command, disp.Dispatch(command, args...)
}

func (disp *StringArgsDispatcher) MustDispatchCombinedCommandAndArgs(commandAndArgs []string) (command string) {
	command, err := disp.DispatchCombinedCommandAndArgs(commandAndArgs)
	if err != nil {
		panic(err)
	}
	return command
}

func (disp *StringArgsDispatcher) PrintCommands(appName string) {
	list := make([]*stringArgsCommand, 0, len(disp.comm))
	for _, cmd := range disp.comm {
		list = append(list, cmd)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].command < list[j].command
	})

	for _, cmd := range list {
		CommandUsageColor.Printf("  %s %s %s\n", appName, cmd.command, cmd.args)
		if cmd.description != "" {
			CommandDescriptionColor.Printf("      %s\n", cmd.description)
		}
		hasAnyArgDesc := false
		for i := 0; i < cmd.args.NumArgs(); i++ {
			if cmd.args.ArgDescription(i) != "" {
				hasAnyArgDesc = true
			}
		}
		if hasAnyArgDesc {
			for i := 0; i < cmd.args.NumArgs(); i++ {
				CommandDescriptionColor.Printf("          <%s:%s> %s\n", cmd.args.ArgName(i), cmd.args.ArgType(i), cmd.args.ArgDescription(i))
			}
		}
		CommandDescriptionColor.Println()
	}
}

func (disp *StringArgsDispatcher) PrintCommandsUsageIntro(appName string, output io.Writer) {
	if len(disp.comm) > 0 {
		fmt.Fprint(output, "Commands:\n")
		disp.PrintCommands(appName)
		fmt.Fprint(output, "Flags:\n")
	}
}
