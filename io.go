package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nekrassov01/mintab"
)

type format int

const (
	formatJSON format = iota
	formatMarkdown
	formatBacklog
)

var formats = []string{
	"json",
	"markdown",
	"backlog",
}

func (f format) String() string {
	if f >= 0 && int(f) < len(formats) {
		return formats[f]
	}
	return ""
}

func fromList(fp string) ([]string, error) {
	if fp == "" {
		return nil, fmt.Errorf("no file provided")
	}
	f, err := os.Open(filepath.Clean(fp))
	if err != nil {
		return nil, fmt.Errorf("cannot open file \"%s\": %w", fp, err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		line, err := checkLine(scanner.Text())
		if err != nil {
			return nil, err
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("no line provided")
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot read line: %w", err)
	}
	return lines, nil
}

func checkLine(line string) (string, error) {
	line = strings.TrimSpace(line)
	if strings.Contains(line, ",") {
		return "", fmt.Errorf("invalid line detected: comma not allowed")
	}
	line = strings.Trim(line, `'"`)
	return line, nil
}

func out(input []*certInfo, format string, omit bool) (string, error) {
	switch format {
	case formatJSON.String():
		return toJSON(input)
	case formatMarkdown.String(), formatBacklog.String():
		return toTable(input, format, omit)
	default:
		return "", fmt.Errorf(
			"cannot parse command line flags: invalid format: allowed values: %s",
			pipeJoin(formats),
		)
	}
}

func toJSON(input []*certInfo) (string, error) {
	b, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", fmt.Errorf("cannot marshal output as json: %w", err)
	}
	return string(b), nil
}

func toTable(input []*certInfo, format string, omit bool) (string, error) {
	var table *mintab.Table
	defaultOpt := mintab.WithEscapeTargets([]string{"*"})
	switch {
	case omit && format == formatMarkdown.String():
		table = mintab.NewTable(defaultOpt, mintab.WithIgnoreFields([]int{8, 9}))
	case omit && format == formatBacklog.String():
		table = mintab.NewTable(defaultOpt, mintab.WithIgnoreFields([]int{8, 9}), mintab.WithFormat(mintab.BacklogFormat))
	case !omit && format == formatMarkdown.String():
		table = mintab.NewTable(defaultOpt)
	case !omit && format == formatBacklog.String():
		table = mintab.NewTable(defaultOpt, mintab.WithFormat(mintab.BacklogFormat))
	}
	if err := table.Load(input); err != nil {
		return "", fmt.Errorf("cannot convert output to table: %w", err)
	}
	return table.Out(), nil
}
