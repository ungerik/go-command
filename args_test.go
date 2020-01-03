package command

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCommandArgsDef struct {
	ArgsDef

	Int0  int    `arg:"int0"`
	Str1  string `arg:"str1"`
	Bool2 bool   `arg:"bool2"`
}

var passedArgsCollector *TestCommandArgsDef

func CommandFunc(int0 int, str1 string, bool2 bool) error {
	// fmt.Println(int0, str1, bool2)
	passedArgsCollector = &TestCommandArgsDef{Int0: int0, Str1: str1, Bool2: bool2}
	return nil
}

func CommandFuncErrResult(int0 int, str1 string, bool2 bool) error {
	passedArgsCollector = &TestCommandArgsDef{Int0: int0, Str1: str1, Bool2: bool2}
	return assert.AnError
}

// func Test_ArgsDef(t *testing.T) {
// 	var commandArgsDef TestCommandArgsDef
// 	stringCommandFunc, err := commandArgsDef.StringArgsFunc(CommandFunc, reflect.TypeOf(commandArgsDef), nil)
// 	assert.NoError(t, err, "Args.StringArgsFunc")
// 	passedArgsCollector = nil
// 	err = stringCommandFunc("123", "Hello World!", "true")
// 	assert.NoError(t, err, "command should return nil")
// 	assert.Equal(t, 123, passedArgsCollector.Int0, "int0")
// 	assert.Equal(t, "Hello World!", passedArgsCollector.Str1, "str1")
// 	assert.Equal(t, true, passedArgsCollector.Bool2, "bool2")

// 	stringCommandFunc, err = commandArgsDef.StringArgsFunc(CommandFunc, reflect.TypeOf(commandArgsDef), nil)
// 	assert.NoError(t, err, "Args.StringArgsFunc")
// 	passedArgsCollector = nil
// 	err = stringCommandFunc("123")
// 	assert.NoError(t, err, "command should return nil")
// 	assert.Equal(t, 123, passedArgsCollector.Int0, "int0")
// 	assert.Equal(t, "", passedArgsCollector.Str1, "str1")
// 	assert.Equal(t, false, passedArgsCollector.Bool2, "bool2")

// 	stringCommandFunc, err = commandArgsDef.StringArgsFunc(CommandFuncErrResult, reflect.TypeOf(commandArgsDef), nil)
// 	assert.NoError(t, err, "Args.StringArgsFunc")
// 	passedArgsCollector = nil
// 	err = stringCommandFunc("123", "Hello World!", "true")
// 	assert.Error(t, err, "command should return an error")
// }

func Test_GetStringArgsFunc(t *testing.T) {
	var commandArgsDef TestCommandArgsDef
	stringCommandFunc, err := GetStringArgsFunc(CommandFunc, &commandArgsDef)
	assert.NoError(t, err, "GetStringArgsFunc")
	passedArgsCollector = nil
	err = stringCommandFunc(context.Background(), "123", "Hello World!", "true")
	assert.NoError(t, err, "command should return nil")
	assert.Equal(t, 123, passedArgsCollector.Int0, "int0")
	assert.Equal(t, "Hello World!", passedArgsCollector.Str1, "str1")
	assert.Equal(t, true, passedArgsCollector.Bool2, "bool2")
}

func Test_GetStringMapArgsFunc(t *testing.T) {
	var commandArgsDef TestCommandArgsDef
	stringCommandFunc, err := GetStringMapArgsFunc(CommandFunc, &commandArgsDef)
	assert.NoError(t, err, "GetStringMapArgsFunc")
	argsMap := map[string]string{
		"int0":  "123",
		"str1":  "Hello World!",
		"bool2": "true",
	}
	passedArgsCollector = nil
	err = stringCommandFunc(context.Background(), argsMap)
	assert.NoError(t, err, "command should return nil")
	assert.Equal(t, 123, passedArgsCollector.Int0, "int0")
	assert.Equal(t, "Hello World!", passedArgsCollector.Str1, "str1")
	assert.Equal(t, true, passedArgsCollector.Bool2, "bool2")
}

type ResultStruct struct {
	ResultCode    int
	ResultMessage string
}

var (
	defaultResultStruct        = ResultStruct{404, "not found"}
	defaultResultStructJSON, _ = json.MarshalIndent(&defaultResultStruct, "", "  ")
)

func CommandFuncStructResult(int0 int, str1 string, bool2 bool) (*ResultStruct, error) {
	passedArgsCollector = &TestCommandArgsDef{Int0: int0, Str1: str1, Bool2: bool2}
	return &defaultResultStruct, nil
}

func Test_WithResultHandler(t *testing.T) {
	var commandArgsDef TestCommandArgsDef
	var resultBuf bytes.Buffer
	stringCommandFunc, err := GetStringArgsFunc(CommandFuncStructResult, &commandArgsDef, PrintTo(&resultBuf))
	assert.NoError(t, err, "GetStringArgsFunc")

	passedArgsCollector = nil
	err = stringCommandFunc(context.Background(), "123", "Hello World!", "true")

	// Check passed args
	assert.Equal(t, 123, passedArgsCollector.Int0, "int0")
	assert.Equal(t, "Hello World!", passedArgsCollector.Str1, "str1")
	assert.Equal(t, true, passedArgsCollector.Bool2, "bool2")

	// Check result struct
	assert.Equal(t, resultBuf.String(), string(defaultResultStructJSON), "equal result JSON")
}
