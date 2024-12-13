package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
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
			args:    []string{appName, insecure, "-d", addr},
			wantErr: false,
		},
		{
			name:    "blank host",
			args:    []string{appName, insecure, "-d", ""},
			wantErr: true,
		},
		{
			name:    "unknown host",
			args:    []string{appName, insecure, "-d", "abc"},
			wantErr: true,
		},
		{
			name:    "list",
			args:    []string{appName, insecure, "-f", filepath.Join("testdata", "1.txt")},
			wantErr: false,
		},
		{
			name:    "list+indent",
			args:    []string{appName, insecure, "-f", filepath.Join("testdata", "2.txt")},
			wantErr: false,
		},
		{
			name:    "list+newline",
			args:    []string{appName, insecure, "-f", filepath.Join("testdata", "3.txt")},
			wantErr: false,
		},
		{
			name:    "list+singleQuote",
			args:    []string{appName, insecure, "-f", filepath.Join("testdata", "4.txt")},
			wantErr: false,
		},
		{
			name:    "list+doubleQuote",
			args:    []string{appName, insecure, "-f", filepath.Join("testdata", "5.txt")},
			wantErr: false,
		},
		{
			name:    "list+comma",
			args:    []string{appName, insecure, "-f", filepath.Join("testdata", "6.txt")},
			wantErr: true,
		},
		{
			name:    "list+blank",
			args:    []string{appName, insecure, "-f", filepath.Join("testdata", "7.txt")},
			wantErr: true,
		},
		{
			name:    "timeout",
			args:    []string{appName, insecure, "-d", addr, "-t", "10s"},
			wantErr: false,
		},
		{
			name:    "timeout invalid string",
			args:    []string{appName, insecure, "-d", addr, "-t", "5"},
			wantErr: true,
		},
		{
			name:    "output json",
			args:    []string{appName, insecure, "-d", addr, "-o", "json"},
			wantErr: false,
		},
		{
			name:    "output markdown",
			args:    []string{appName, insecure, "-d", addr, "-o", "markdown"},
			wantErr: false,
		},
		{
			name:    "output backlog",
			args:    []string{appName, insecure, "-d", addr, "-o", "backlog"},
			wantErr: false,
		},
		{
			name:    "output unknown format",
			args:    []string{appName, insecure, "-d", addr, "-o", "unknown"},
			wantErr: true,
		},
		{
			name:    "no timeinfo",
			args:    []string{appName, insecure, "-d", addr, "-n"},
			wantErr: false,
		},
		{
			name:    "timezone",
			args:    []string{appName, insecure, "-d", addr, "-z", "UTC"},
			wantErr: false,
		},
		{
			name:    "completion bash",
			args:    []string{appName, "-c", "bash"},
			wantErr: false,
		},
		{
			name:    "completion zsh",
			args:    []string{appName, "-c", "zsh"},
			wantErr: false,
		},
		{
			name:    "completion pwsh",
			args:    []string{appName, "-c", "pwsh"},
			wantErr: false,
		},
		{
			name:    "completion unsupported",
			args:    []string{appName, "-c", "fish"},
			wantErr: true,
		},
		{
			name:    "unknown flag provided",
			args:    []string{appName, "-1"},
			wantErr: true,
		},
		{
			name:    "no flag provided",
			args:    []string{appName},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := os.Setenv(canonicalName+"_NON_INTERACTIVE", "true")
		if err != nil {
			t.Fatal(err)
		}
		t.Run(tt.name, func(t *testing.T) {
			err := newApp(io.Discard).RunContext(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
