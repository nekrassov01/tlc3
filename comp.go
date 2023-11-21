package main

import (
	_ "embed"
	"fmt"
)

//go:embed completion/tlc3.bash
var bashCompletion string

//go:embed completion/tlc3.zsh
var zshCompletion string

//go:embed completion/tlc3.ps1
var pwshCompletion string

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

func comp(s string) error {
	switch s {
	case bash.String():
		fmt.Println(bashCompletion)
	case zsh.String():
		fmt.Println(zshCompletion)
	case pwsh.String():
		fmt.Println(pwshCompletion)
	default:
		return fmt.Errorf(
			"cannot parse command line flags: invalid completion shell: allowed values: %s",
			pipeJoin(shells),
		)
	}
	return nil
}
