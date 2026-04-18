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
			name:    "list+comma",
			args:    []string{name, insecure, "-f", filepath.Join("..", "testdata", "1.txt")},
			wantErr: true,
		},
		{
			name:    "list+blank",
			args:    []string{name, insecure, "-f", filepath.Join("..", "testdata", "2.txt")},
			wantErr: true,
		},
		{
			name:    "timeout invalid string",
			args:    []string{name, insecure, "-a", "example.com", "-t", "5"},
			wantErr: true,
		},
		{
			name:    "output unknown format",
			args:    []string{name, insecure, "-a", "example.com", "-o", "unknown"},
			wantErr: true,
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
