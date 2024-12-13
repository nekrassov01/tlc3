package main

import (
	_ "embed"
	"fmt"
	"io"
)

//go:embed completions/tlc3.bash
var completionBash string

//go:embed completions/tlc3.zsh
var completionZsh string

//go:embed completions/tlc3.ps1
var completionPwsh string

type shell int

const (
	bash shell = iota
	zsh
	pwsh
)

var shells = []string{
	"bash",
	"zsh",
	"pwsh",
}

func (c shell) String() string {
	if c >= 0 && int(c) < len(shells) {
		return shells[c]
	}
	return ""
}

func comp(w io.Writer, s string) error {
	switch s {
	case bash.String():
		fmt.Fprintln(w, completionBash)
	case zsh.String():
		fmt.Fprintln(w, completionZsh)
	case pwsh.String():
		fmt.Fprintln(w, completionPwsh)
	default:
		return fmt.Errorf("invalid completion shell: allowed values: %s", pipeJoin(shells))
	}
	return nil
}
