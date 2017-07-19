package command

import (
	"errors"
	"fmt"
	"sort"
)

var ErrNotFound = errors.New("command not found")

const Default = ""

type stringArgsCommand struct {
	command         string
	description     string
	args            Args
	commandFunc     interface{}
	stringArgsFunc  StringArgsFunc
	resultsHandlers []ResultsHandler
}

type StringArgsDispatcher map[string]*stringArgsCommand

func (disp StringArgsDispatcher) AddCommand(command, description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) error {
	stringArgsFunc, err := GetStringArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		return err
	}
	disp[command] = &stringArgsCommand{
		command:         command,
		description:     description,
		args:            args,
		commandFunc:     commandFunc,
		stringArgsFunc:  stringArgsFunc,
		resultsHandlers: resultsHandlers,
	}
	return nil
}

func (disp StringArgsDispatcher) MustAddCommand(command, description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) {
	err := disp.AddCommand(command, description, commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
}

func (disp StringArgsDispatcher) AddDefaultCommand(description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) error {
	stringArgsFunc, err := GetStringArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		return err
	}
	disp[Default] = &stringArgsCommand{
		command:         Default,
		description:     description,
		args:            args,
		commandFunc:     commandFunc,
		stringArgsFunc:  stringArgsFunc,
		resultsHandlers: resultsHandlers,
	}
	return nil
}

func (disp StringArgsDispatcher) MustAddDefaultCommand(description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) {
	err := disp.AddDefaultCommand(description, commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
}

func (disp StringArgsDispatcher) HasCommnd(command string) bool {
	_, found := disp[command]
	return found
}

func (disp StringArgsDispatcher) HasDefaultCommnd() bool {
	_, found := disp[Default]
	return found
}

func (disp StringArgsDispatcher) Dispatch(command string, args ...string) error {
	cmd, found := disp[command]
	if !found {
		return ErrNotFound
	}
	return cmd.stringArgsFunc(args...)
}

func (disp StringArgsDispatcher) MustDispatch(command string, args ...string) {
	err := disp.Dispatch(command, args...)
	if err != nil {
		panic(err)
	}
}

func (disp StringArgsDispatcher) DispatchDefaultCommand() error {
	return disp.Dispatch(Default)
}

func (disp StringArgsDispatcher) MustDispatchDefaultCommand() {
	err := disp.DispatchDefaultCommand()
	if err != nil {
		panic(err)
	}
}

func (disp StringArgsDispatcher) DispatchCombined(commandAndArgs []string) (command string, err error) {
	if len(commandAndArgs) == 0 {
		return Default, disp.DispatchDefaultCommand()
	}
	command = commandAndArgs[0]
	args := commandAndArgs[1:]
	return command, disp.Dispatch(command, args...)
}

func (disp StringArgsDispatcher) MustDispatchCombined(commandAndArgs []string) (command string) {
	command, err := disp.DispatchCombined(commandAndArgs)
	if err != nil {
		panic(err)
	}
	return command
}

func (disp StringArgsDispatcher) PrintCommands(appName string) {
	list := make([]*stringArgsCommand, 0, len(disp))
	for _, cmd := range disp {
		list = append(list, cmd)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].command < list[j].command
	})

	for _, cmd := range list {
		fmt.Printf("  %s %s %s\n", appName, cmd.command, cmd.args)
		if len(cmd.description) == 0 {
			fmt.Println()
		} else {
			fmt.Printf("      %s\n", cmd.description)
		}
	}
}
