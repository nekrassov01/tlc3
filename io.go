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
	formatTextTable
	formatMarkdownTable
	formatBacklogTable
)

var formats = []string{
	"json",
	"table",
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
	case formatTextTable.String(), formatMarkdownTable.String(), formatBacklogTable.String():
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
	ignore := mintab.WithIgnoreFields([]int{8, 9})
	escape := mintab.WithEscapeTargets([]string{"*"})
	markdown := mintab.WithFormat(mintab.MarkdownFormat)
	backlog := mintab.WithFormat(mintab.BacklogFormat)
	switch {
	case omit && format == formatTextTable.String():
		table = mintab.NewTable(ignore)
	case omit && format == formatMarkdownTable.String():
		table = mintab.NewTable(ignore, escape, markdown)
	case omit && format == formatBacklogTable.String():
		table = mintab.NewTable(ignore, escape, backlog)
	case !omit && format == formatTextTable.String():
		table = mintab.NewTable()
	case !omit && format == formatMarkdownTable.String():
		table = mintab.NewTable(escape, markdown)
	case !omit && format == formatBacklogTable.String():
		table = mintab.NewTable(escape, backlog)
	}
	if err := table.Load(input); err != nil {
		return "", fmt.Errorf("cannot convert output to table: %w", err)
	}
	return table.Out(), nil
}
