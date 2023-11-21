package main

import (
	_ "embed"
	"testing"
)

func Test_comp(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "bash",
			args: args{
				s: bash.String(),
			},
			wantErr: false,
		},
		{
			name: "zsh",
			args: args{
				s: zsh.String(),
			},
			wantErr: false,
		},
		{
			name: "pwsh",
			args: args{
				s: pwsh.String(),
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				s: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := comp(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("comp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
