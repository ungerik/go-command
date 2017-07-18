package command

import "errors"

var ErrCommandNotFound = errors.New("command not found")

const Default = ""

type StringArgsDispatcher map[string]StringArgsFunc

func (disp StringArgsDispatcher) AddCommand(command string, args Args, commandFunc interface{}, resultsHandler ResultsHandler) error {
	stringArgsFunc, err := GetStringArgsFunc(args, commandFunc, resultsHandler)
	if err != nil {
		return err
	}
	disp[command] = stringArgsFunc
	return nil
}

func (disp StringArgsDispatcher) MustAddCommand(command string, args Args, commandFunc interface{}, resultsHandler ResultsHandler) {
	err := disp.AddCommand(command, args, commandFunc, resultsHandler)
	if err != nil {
		panic(err)
	}
}

func (disp StringArgsDispatcher) AddDefaultCommand(args Args, commandFunc interface{}, resultsHandler ResultsHandler) error {
	stringArgsFunc, err := GetStringArgsFunc(args, commandFunc, resultsHandler)
	if err != nil {
		return err
	}
	disp[Default] = stringArgsFunc
	return nil
}

func (disp StringArgsDispatcher) MustAddDefaultCommand(args Args, commandFunc interface{}, resultsHandler ResultsHandler) {
	err := disp.AddDefaultCommand(args, commandFunc, resultsHandler)
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
	commandFunc, found := disp[command]
	if !found {
		return ErrCommandNotFound
	}
	return commandFunc(args...)
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

func (disp StringArgsDispatcher) DispatchCombined(commandAndArgs []string) error {
	if len(commandAndArgs) == 0 {
		return disp.DispatchDefaultCommand()
	}
	return disp.Dispatch(commandAndArgs[0], commandAndArgs[1:]...)
}

func (disp StringArgsDispatcher) MustDispatchCombined(commandAndArgs []string) {
	err := disp.DispatchCombined(commandAndArgs)
	if err != nil {
		panic(err)
	}
}
