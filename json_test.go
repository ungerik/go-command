package command

import (
	"reflect"
	"testing"
)

func Test_unmarshalJSONFunctionArgs(t *testing.T) {
	funcInfo := func(f interface{}) FunctionInfo {
		info, err := ReflectFunctionInfo("f", f)
		if err != nil {
			panic(err)
		}
		return info
	}
	type args struct {
		f          FunctionInfo
		jsonObject []byte
	}
	tests := []struct {
		name     string
		args     args
		wantArgs []interface{}
		wantErr  bool
	}{
		{
			name: "empty",
			args: args{
				f:          funcInfo(func() {}),
				jsonObject: []byte(`{"a0": "ignored"}`),
			},
			wantArgs: []interface{}{},
		},
		{
			name: "default 0",
			args: args{
				f:          funcInfo(func(string, int) {}),
				jsonObject: []byte(`{"a0": "default"}`),
			},
			wantArgs: []interface{}{"default", 0},
		},
		{
			name: "hello 666",
			args: args{
				f:          funcInfo(func(string, int) {}),
				jsonObject: []byte(`{"a0": "hello", "a1": 666, "a2": "ignored"}`),
			},
			wantArgs: []interface{}{"hello", 666},
		},
		{
			name: "ptr",
			args: args{
				f:          funcInfo(func(*string, *string, interface{}) {}),
				jsonObject: []byte(`{"a0": "", "a2": null}`),
			},
			wantArgs: []interface{}{new(string), (*string)(nil), nil},
		},
		// wantErr
		{
			name: "JSON array",
			args: args{
				f:          funcInfo(func() {}),
				jsonObject: []byte(`[]`),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArgs, err := unmarshalJSONFunctionArgs(tt.args.f, tt.args.jsonObject)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalJSONFunctionArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("unmarshalJSONFunctionArgs() = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
