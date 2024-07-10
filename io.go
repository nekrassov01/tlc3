package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

func out(in []*certInfo, out io.Writer, format string, omit bool) error {
	switch format {
	case formatJSON.String():
		return toJSON(in, out)
	case formatTextTable.String(), formatMarkdownTable.String(), formatBacklogTable.String():
		return toTable(in, out, format, omit)
	default:
		return fmt.Errorf(
			"cannot parse command line flags: invalid format: allowed values: %s",
			pipeJoin(formats),
		)
	}
}

func toJSON(in []*certInfo, out io.Writer) error {
	b := json.NewEncoder(out)
	b.SetIndent("", "  ")
	if err := b.Encode(in); err != nil {
		return fmt.Errorf("cannot marshal output as json: %w", err)
	}
	return nil
}

func toTable(in []*certInfo, out io.Writer, format string, omit bool) error {
	opts := make([]mintab.Option, 0, 2)
	switch format {
	case formatTextTable.String():
	case formatMarkdownTable.String():
		opts = append(opts, mintab.WithFormat(mintab.FormatMarkdown))
	case formatBacklogTable.String():
		opts = append(opts, mintab.WithFormat(mintab.FormatBacklog))
	}
	if omit {
		opts = append(opts, mintab.WithIgnoreFields([]int{8, 9}))
	}
	table := mintab.New(out, opts...)
	if err := table.Load(in); err != nil {
		return fmt.Errorf("cannot convert output to table: %w", err)
	}
	table.Out()
	return nil
}
