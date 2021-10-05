package command

import (
	"testing"
)

func f0()    {}
func f1(int) {}

func TestGenerateFunctionTODO(t *testing.T) {
	tests := []struct {
		name    string
		f       interface{}
		wantErr string
	}{
		{
			name:    "f0",
			f:       f0,
			wantErr: "GenerateFunctionTODO(func())",
		},
		{
			name:    "f1",
			f:       f1,
			wantErr: "GenerateFunctionTODO(func(int))",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFunctionTODO(tt.f)
			if gotErr, ok := got.(errFunction); !ok || gotErr.err.Error() != tt.wantErr {
				t.Errorf("GenerateFunctionTODO() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}
