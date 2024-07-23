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

func out(infos []*certInfo, w io.Writer, format string, omit bool) error {
	switch format {
	case formatJSON.String():
		return toJSON(infos, w)
	case formatTextTable.String(), formatMarkdownTable.String(), formatBacklogTable.String():
		return toTable(infos, w, format, omit)
	default:
		return fmt.Errorf(
			"cannot parse command line flags: invalid format: allowed values: %s",
			pipeJoin(formats),
		)
	}
}

func toJSON(infos []*certInfo, w io.Writer) error {
	b := json.NewEncoder(w)
	b.SetIndent("", "  ")
	if err := b.Encode(infos); err != nil {
		return fmt.Errorf("cannot marshal output as json: %w", err)
	}
	return nil
}

func toTable(infos []*certInfo, w io.Writer, format string, omit bool) error {
	opts := make([]mintab.Option, 0, 2)
	switch format {
	case formatTextTable.String():
	case formatMarkdownTable.String():
		opts = append(opts, mintab.WithFormat(mintab.MarkdownFormat))
	case formatBacklogTable.String():
		opts = append(opts, mintab.WithFormat(mintab.BacklogFormat))
	}
	if omit {
		opts = append(opts, mintab.WithIgnoreFields([]int{8, 9}))
	}
	table := mintab.New(w, opts...)
	if err := table.Load(toInput(infos)); err != nil {
		return fmt.Errorf("cannot convert output to table: %w", err)
	}
	table.Render()
	return nil
}

func toInput(infos []*certInfo) mintab.Input {
	header := []string{
		"DomainName",
		"AccessPort",
		"IPAddresses",
		"Issuer",
		"CommonName",
		"SANs",
		"NotBefore",
		"NotAfter",
		"CurrentTime",
		"DaysLeft",
	}
	data := make([][]any, len(infos))
	for i, info := range infos {
		data[i] = []any{
			info.DomainName,
			info.AccessPort,
			info.IPAddresses,
			info.Issuer,
			info.CommonName,
			info.SANs,
			info.NotBefore,
			info.NotAfter,
			info.CurrentTime,
			info.DaysLeft,
		}
	}
	return mintab.Input{
		Header: header,
		Data:   data,
	}
}
