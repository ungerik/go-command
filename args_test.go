package command

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCommandArgsDef struct {
	ArgsDef
	Int0  int    `cmd:"int0"`
	Str1  string `cmd:"str1"`
	Bool2 bool   `cmd:"bool2"`
}

var passedArgs *TestCommandArgsDef

func CommandFunc(int0 int, str1 string, bool2 bool) error {
	// fmt.Println(int0, str1, bool2)
	passedArgs = &TestCommandArgsDef{Int0: int0, Str1: str1, Bool2: bool2}
	return nil
}

func CommandFuncErr(int0 int, str1 string, bool2 bool) error {
	passedArgs = &TestCommandArgsDef{Int0: int0, Str1: str1, Bool2: bool2}
	return assert.AnError
}

func Test_ArgsDef(t *testing.T) {
	var commandArgsDef TestCommandArgsDef
	stringCommandFunc, err := commandArgsDef.StringArgsFunc(reflect.TypeOf(commandArgsDef), CommandFunc, nil)
	assert.NoError(t, err, "Args.StringArgsFunc")
	passedArgs = nil
	err = stringCommandFunc("123", "Hello World!", "true")
	assert.NoError(t, err, "command should return nil")
	assert.Equal(t, 123, passedArgs.Int0, "int0")
	assert.Equal(t, "Hello World!", passedArgs.Str1, "str1")
	assert.Equal(t, true, passedArgs.Bool2, "bool2")

	stringCommandFunc, err = commandArgsDef.StringArgsFunc(reflect.TypeOf(commandArgsDef), CommandFunc, nil)
	assert.NoError(t, err, "Args.StringArgsFunc")
	passedArgs = nil
	err = stringCommandFunc("123")
	assert.NoError(t, err, "command should return nil")
	assert.Equal(t, 123, passedArgs.Int0, "int0")
	assert.Equal(t, "", passedArgs.Str1, "str1")
	assert.Equal(t, false, passedArgs.Bool2, "bool2")

	stringCommandFunc, err = commandArgsDef.StringArgsFunc(reflect.TypeOf(commandArgsDef), CommandFuncErr, nil)
	assert.NoError(t, err, "Args.StringArgsFunc")
	passedArgs = nil
	err = stringCommandFunc("123", "Hello World!", "true")
	assert.Error(t, err, "command should return an error")
}

func Test_GetStringArgsFunc(t *testing.T) {
	var commandArgsDef TestCommandArgsDef
	stringCommandFunc, err := GetStringArgsFunc(&commandArgsDef, CommandFunc)
	assert.NoError(t, err, "GetStringArgsFunc")
	passedArgs = nil
	err = stringCommandFunc("123", "Hello World!", "true")
	assert.NoError(t, err, "command should return nil")
	assert.Equal(t, 123, passedArgs.Int0, "int0")
	assert.Equal(t, "Hello World!", passedArgs.Str1, "str1")
	assert.Equal(t, true, passedArgs.Bool2, "bool2")
}

func Test_GetStringMapArgsFunc(t *testing.T) {
	var commandArgsDef TestCommandArgsDef
	stringCommandFunc, err := GetStringMapArgsFunc(&commandArgsDef, CommandFunc)
	assert.NoError(t, err, "GetStringMapArgsFunc")
	argsMap := map[string]string{
		"int0":  "123",
		"str1":  "Hello World!",
		"bool2": "true",
	}
	passedArgs = nil
	err = stringCommandFunc(argsMap)
	assert.NoError(t, err, "command should return nil")
	assert.Equal(t, 123, passedArgs.Int0, "int0")
	assert.Equal(t, "Hello World!", passedArgs.Str1, "str1")
	assert.Equal(t, true, passedArgs.Bool2, "bool2")
}
