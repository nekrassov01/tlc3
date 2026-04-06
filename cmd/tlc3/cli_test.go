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
			args:    []string{name, insecure, "-a", "example.com"},
			wantErr: false,
		},
		{
			name:    "blank host",
			args:    []string{name, insecure, "-a", ""},
			wantErr: true,
		},
		{
			name:    "unknown host",
			args:    []string{name, insecure, "-a", "abc"},
			wantErr: true,
		},
		{
			name:    "list",
			args:    []string{name, insecure, "-f", filepath.Join("testdata", "1.txt")},
			wantErr: false,
		},
		{
			name:    "list+indent",
			args:    []string{name, insecure, "-f", filepath.Join("testdata", "2.txt")},
			wantErr: false,
		},
		{
			name:    "list+newline",
			args:    []string{name, insecure, "-f", filepath.Join("testdata", "3.txt")},
			wantErr: false,
		},
		{
			name:    "list+singleQuote",
			args:    []string{name, insecure, "-f", filepath.Join("testdata", "4.txt")},
			wantErr: false,
		},
		{
			name:    "list+doubleQuote",
			args:    []string{name, insecure, "-f", filepath.Join("testdata", "5.txt")},
			wantErr: false,
		},
		{
			name:    "list+comma",
			args:    []string{name, insecure, "-f", filepath.Join("testdata", "6.txt")},
			wantErr: true,
		},
		{
			name:    "list+blank",
			args:    []string{name, insecure, "-f", filepath.Join("testdata", "7.txt")},
			wantErr: true,
		},
		{
			name:    "timeout",
			args:    []string{name, insecure, "-a", "example.com", "-t", "10s"},
			wantErr: false,
		},
		{
			name:    "timeout invalid string",
			args:    []string{name, insecure, "-a", "example.com", "-t", "5"},
			wantErr: true,
		},
		{
			name:    "output json",
			args:    []string{name, insecure, "-a", "example.com", "-o", "json"},
			wantErr: false,
		},
		{
			name:    "output markdown",
			args:    []string{name, insecure, "-a", "example.com", "-o", "markdown"},
			wantErr: false,
		},
		{
			name:    "output backlog",
			args:    []string{name, insecure, "-a", "example.com", "-o", "backlog"},
			wantErr: false,
		},
		{
			name:    "output unknown format",
			args:    []string{name, insecure, "-a", "example.com", "-o", "unknown"},
			wantErr: true,
		},
		{
			name:    "static",
			args:    []string{name, insecure, "-a", "example.com", "-s"},
			wantErr: false,
		},
		{
			name:    "timezone",
			args:    []string{name, insecure, "-a", "example.com", "-z", "UTC"},
			wantErr: false,
		},
		{
			name:    "unknown flag provided",
			args:    []string{name, "-1"},
			wantErr: true,
		},
		{
			name:    "no flag provided",
			args:    []string{name},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		err := os.Setenv("TLC3_NON_INTERACTIVE", "true")
		if err != nil {
			t.Fatal(err)
		}
		t.Run(tt.name, func(t *testing.T) {
			err := newCmd(io.Discard, io.Discard).Run(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
