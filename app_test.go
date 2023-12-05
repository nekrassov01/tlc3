package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_cli(t *testing.T) {
	insecure := "-i"
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "basic",
			args:    []string{Name, insecure, "-d", addr},
			wantErr: false,
		},
		{
			name:    "blank host",
			args:    []string{Name, insecure, "-d", ""},
			wantErr: true,
		},
		{
			name:    "unknown host",
			args:    []string{Name, insecure, "-d", "localhost"},
			wantErr: true,
		},
		{
			name:    "list",
			args:    []string{Name, insecure, "-l", filepath.Join("testdata", "1.txt")},
			wantErr: false,
		},
		{
			name:    "list+indent",
			args:    []string{Name, insecure, "-l", filepath.Join("testdata", "2.txt")},
			wantErr: false,
		},
		{
			name:    "list+newline",
			args:    []string{Name, insecure, "-l", filepath.Join("testdata", "3.txt")},
			wantErr: false,
		},
		{
			name:    "list+singleQuote",
			args:    []string{Name, insecure, "-l", filepath.Join("testdata", "4.txt")},
			wantErr: false,
		},
		{
			name:    "list+doubleQuote",
			args:    []string{Name, insecure, "-l", filepath.Join("testdata", "5.txt")},
			wantErr: false,
		},
		{
			name:    "list+comma",
			args:    []string{Name, insecure, "-l", filepath.Join("testdata", "6.txt")},
			wantErr: true,
		},
		{
			name:    "list+blank",
			args:    []string{Name, insecure, "-l", filepath.Join("testdata", "7.txt")},
			wantErr: true,
		},
		{
			name:    "timeout",
			args:    []string{Name, insecure, "-d", addr, "-t", "10s"},
			wantErr: false,
		},
		{
			name:    "timeout invalid string",
			args:    []string{Name, insecure, "-d", addr, "-t", "5"},
			wantErr: true,
		},
		{
			name:    "output json",
			args:    []string{Name, insecure, "-d", addr, "-o", "json"},
			wantErr: false,
		},
		{
			name:    "output markdown",
			args:    []string{Name, insecure, "-d", addr, "-o", "markdown"},
			wantErr: false,
		},
		{
			name:    "output backlog",
			args:    []string{Name, insecure, "-d", addr, "-o", "backlog"},
			wantErr: false,
		},
		{
			name:    "output unknown format",
			args:    []string{Name, insecure, "-d", addr, "-o", "unknown"},
			wantErr: true,
		},
		{
			name:    "no timeinfo",
			args:    []string{Name, insecure, "-d", addr, "-n"},
			wantErr: false,
		},
		{
			name:    "completion bash",
			args:    []string{Name, "-c", "bash"},
			wantErr: false,
		},
		{
			name:    "completion zsh",
			args:    []string{Name, "-c", "zsh"},
			wantErr: false,
		},
		{
			name:    "completion pwsh",
			args:    []string{Name, "-c", "pwsh"},
			wantErr: false,
		},
		{
			name:    "completion unsupported",
			args:    []string{Name, "-c", "fish"},
			wantErr: true,
		},
		{
			name:    "unknown flag provided",
			args:    []string{Name, "-1"},
			wantErr: true,
		},
		{
			name:    "no flag provided",
			args:    []string{Name},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		err := os.Setenv(strings.ToUpper(Name)+"_NON_INTERACTIVE", "true")
		if err != nil {
			t.Fatal(err)
		}
		t.Run(tt.name, func(t *testing.T) {
			err := newApp().cli.RunContext(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
