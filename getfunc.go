package command

func GetStringArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) (StringArgsFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringArgsFunc(commandFunc, resultsHandlers)
}

func MustGetStringArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) StringArgsFunc {
	f, err := GetStringArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func GetStringMapArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) (StringMapArgsFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringMapArgsFunc(commandFunc, resultsHandlers)
}

func MustGetStringMapArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) StringMapArgsFunc {
	f, err := GetStringMapArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func GetMapArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) (MapArgsFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.MapArgsFunc(commandFunc, resultsHandlers)
}

func MustGetMapArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) MapArgsFunc {
	f, err := GetMapArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func GetJSONArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) (JSONArgsFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.JSONArgsFunc(commandFunc, resultsHandlers)
}

func MustGetJSONArgsFunc(commandFunc interface{}, args Args, resultsHandlers ...ResultsHandler) JSONArgsFunc {
	f, err := GetJSONArgsFunc(commandFunc, args, resultsHandlers...)
	if err != nil {
		panic(err)
	}
	return f
}

func GetStringArgsResultValuesFunc(commandFunc interface{}, args Args) (StringArgsResultValuesFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringArgsResultValuesFunc(commandFunc)
}

func MustGetStringArgsResultValuesFunc(commandFunc interface{}, args Args) StringArgsResultValuesFunc {
	f, err := GetStringArgsResultValuesFunc(commandFunc, args)
	if err != nil {
		panic(err)
	}
	return f
}

func GetStringMapArgsResultValuesFunc(commandFunc interface{}, args Args) (StringMapArgsResultValuesFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.StringMapArgsResultValuesFunc(commandFunc)
}

func MustGetStringMapArgsResultValuesFunc(commandFunc interface{}, args Args) StringMapArgsResultValuesFunc {
	f, err := GetStringMapArgsResultValuesFunc(commandFunc, args)
	if err != nil {
		panic(err)
	}
	return f
}

func GetMapArgsResultValuesFunc(commandFunc interface{}, args Args) (MapArgsResultValuesFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.MapArgsResultValuesFunc(commandFunc)
}

func MustGetMapArgsResultValuesFunc(commandFunc interface{}, args Args) MapArgsResultValuesFunc {
	f, err := GetMapArgsResultValuesFunc(commandFunc, args)
	if err != nil {
		panic(err)
	}
	return f
}

func GetJSONArgsResultValuesFunc(commandFunc interface{}, args Args) (JSONArgsResultValuesFunc, error) {
	// Note: here happens something unexpected!
	// args implements the Args interface with ArgsDef.
	// This looks like a virtual method call, but of course it is not.
	// The first args is interpreted as (*ArgsDef) to do the method call.
	// We can't use that to get the type that embedds ArgsDef,
	// because ArgsDef knows nothing about the outer embedding type.
	// But args, the first argument to the method, has all the type information,
	// because here the complete outer embedding struct is passed.
	argsImpl := args.(ArgsImpl)
	err := argsImpl.Init(args)
	if err != nil {
		return nil, err
	}
	return argsImpl.JSONArgsResultValuesFunc(commandFunc)
}

func MustGetJSONArgsResultValuesFunc(commandFunc interface{}, args Args) JSONArgsResultValuesFunc {
	f, err := GetJSONArgsResultValuesFunc(commandFunc, args)
	if err != nil {
		panic(err)
	}
	return f
}
