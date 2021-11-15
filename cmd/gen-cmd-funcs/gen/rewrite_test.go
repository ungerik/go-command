package gen

import "testing"

func Test_parseImplementsComment(t *testing.T) {
	type args struct {
		implementor string
		comment     string
	}
	tests := []struct {
		name            string
		args            args
		wantWrappedFunc string
		wantImplements  Impl
		wantErr         bool
	}{
		{
			name:            "function.Wrapper (generated code)",
			args:            args{implementor: "myFunction", comment: "myFunction wraps my.Function as function.Wrapper (generated code)"},
			wantWrappedFunc: "my.Function",
			wantImplements:  ImplWrapper,
		},
		{
			name:            "function.Wrapper",
			args:            args{implementor: "myFunction", comment: " myFunction wraps my.Function as function.Wrapper "},
			wantWrappedFunc: "my.Function",
			wantImplements:  ImplWrapper,
		},
		{
			name:            "function.Description",
			args:            args{implementor: "myFunction", comment: "myFunction wraps MyFunction as function.Description (generated code)"},
			wantWrappedFunc: "MyFunction",
			wantImplements:  ImplDescription,
		},

		// Invalid:
		{
			name:    "empty",
			args:    args{implementor: "", comment: ""},
			wantErr: true,
		},
		{
			name:    "missing wrapped func",
			args:    args{implementor: "myFunction", comment: "myFunction wraps as function.Wrapper"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWrappedFunc, gotImplements, err := parseImplementsComment(tt.args.implementor, tt.args.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseImplementsComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWrappedFunc != tt.wantWrappedFunc {
				t.Errorf("parseImplementsComment() gotWrappedFunc = %v, want %v", gotWrappedFunc, tt.wantWrappedFunc)
			}
			if gotImplements != tt.wantImplements {
				t.Errorf("parseImplementsComment() gotImplements = %v, want %v", gotImplements, tt.wantImplements)
			}
		})
	}
}
