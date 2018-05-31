package command

import (
	"fmt"
	"io"
	"sort"

	"github.com/domonda/errors"
)

type SuperCommandNotFound string

func (s SuperCommandNotFound) Error() string {
	return fmt.Sprintf("Super command '%s' not found", string(s))
}

type SuperStringArgsDispatcher struct {
	sub     map[string]*StringArgsDispatcher
	loggers []StringArgsCommandLogger
}

func NewSuperStringArgsDispatcher(loggers ...StringArgsCommandLogger) *SuperStringArgsDispatcher {
	return &SuperStringArgsDispatcher{
		sub:     make(map[string]*StringArgsDispatcher),
		loggers: loggers,
	}
}

func (disp *SuperStringArgsDispatcher) AddSuperCommand(superCommand string) (subDisp *StringArgsDispatcher, err error) {
	if superCommand != "" {
		if err := checkCommandChars(superCommand); err != nil {
			return nil, errors.Wrapf(err, "Command '%s'", superCommand)
		}
	}
	if _, exists := disp.sub[superCommand]; exists {
		return nil, errors.Errorf("Super command already added: '%s'", superCommand)
	}
	subDisp = NewStringArgsDispatcher(disp.loggers...)
	disp.sub[superCommand] = subDisp
	return subDisp, nil
}

func (disp *SuperStringArgsDispatcher) MustAddSuperCommand(superCommand string) (subDisp *StringArgsDispatcher) {
	subDisp, err := disp.AddSuperCommand(superCommand)
	if err != nil {
		panic(errors.Wrapf(err, "MustAddSuperCommand(%s)", superCommand))
	}
	return subDisp
}

func (disp *SuperStringArgsDispatcher) AddDefaultCommand(description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) error {
	subDisp, err := disp.AddSuperCommand(Default)
	if err != nil {
		return err
	}
	return subDisp.AddDefaultCommand(description, commandFunc, args, resultsHandlers...)
}

func (disp *SuperStringArgsDispatcher) MustAddDefaultCommand(description string, commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) {
	err := disp.AddDefaultCommand(description, commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(errors.Wrapf(err, "MustAddDefaultCommand(%s)", description))
	}
}
func (disp *SuperStringArgsDispatcher) HasCommnd(superCommand string) bool {
	sub, ok := disp.sub[superCommand]
	if !ok {
		return false
	}
	return sub.HasDefaultCommnd()
}

func (disp *SuperStringArgsDispatcher) HasSubCommnd(superCommand, command string) bool {
	sub, ok := disp.sub[superCommand]
	if !ok {
		return false
	}
	return sub.HasCommnd(command)
}

func (disp *SuperStringArgsDispatcher) Dispatch(superCommand, command string, args ...string) error {
	sub, ok := disp.sub[superCommand]
	if !ok {
		return SuperCommandNotFound(superCommand)
	}
	return sub.Dispatch(command, args...)
}

func (disp *SuperStringArgsDispatcher) MustDispatch(superCommand, command string, args ...string) {
	err := disp.Dispatch(superCommand, command, args...)
	if err != nil {
		panic(errors.Wrapf(err, "Command '%s'", command))
	}
}

func (disp *SuperStringArgsDispatcher) DispatchDefaultCommand() error {
	return disp.Dispatch(Default, Default)
}

func (disp *SuperStringArgsDispatcher) MustDispatchDefaultCommand() {
	err := disp.DispatchDefaultCommand()
	if err != nil {
		panic(errors.Wrap(err, "Default command"))
	}
}

func (disp *SuperStringArgsDispatcher) DispatchCombinedCommandAndArgs(commandAndArgs []string) (superCommand, command string, err error) {
	var args []string
	switch len(commandAndArgs) {
	case 0:
		superCommand = Default
		command = Default
	case 1:
		superCommand = commandAndArgs[0]
		command = Default
	default:
		superCommand = commandAndArgs[0]
		sub, ok := disp.sub[superCommand]
		if ok && sub.HasDefaultCommnd() {
			command = Default
			args = commandAndArgs[1:]
		} else {
			command = commandAndArgs[1]
			args = commandAndArgs[2:]
		}
	}
	return superCommand, command, disp.Dispatch(superCommand, command, args...)
}

func (disp *SuperStringArgsDispatcher) MustDispatchCombinedCommandAndArgs(commandAndArgs []string) (superCommand, command string) {
	superCommand, command, err := disp.DispatchCombinedCommandAndArgs(commandAndArgs)
	if err != nil {
		panic(errors.Wrapf(err, "MustDispatchCombinedCommandAndArgs(%v)", commandAndArgs))
	}
	return superCommand, command
}

func (disp *SuperStringArgsDispatcher) PrintCommands(appName string) {
	type superCmd struct {
		super string
		cmd   *stringArgsCommand
	}

	var list []superCmd
	for super, sub := range disp.sub {
		for _, cmd := range sub.comm {
			list = append(list, superCmd{super: super, cmd: cmd})
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].super == list[j].super {
			return list[i].cmd.command < list[j].cmd.command
		}
		return list[i].super < list[j].super
	})

	for i := range list {
		cmd := list[i].cmd
		command := list[i].super
		if cmd.command != Default {
			command += " " + cmd.command
		}

		CommandUsageColor.Printf("  %s %s %s\n", appName, command, cmd.args)
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

func (disp *SuperStringArgsDispatcher) PrintCommandsUsageIntro(appName string, output io.Writer) {
	if len(disp.sub) > 0 {
		fmt.Fprint(output, "Commands:\n")
		disp.PrintCommands(appName)
		fmt.Fprint(output, "Flags:\n")
	}
}
