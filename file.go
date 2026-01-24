package tlc3

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetDomainsFromFile(fp string) ([]string, error) {
	if fp == "" {
		return nil, errors.New("no file provided")
	}
	f, err := os.Open(filepath.Clean(fp))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		line, err := sanitize(scanner.Text())
		if err != nil {
			return nil, err
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("no line provided: %s", fp)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func sanitize(line string) (string, error) {
	line = strings.TrimSpace(line)
	if strings.Contains(line, ",") {
		return "", fmt.Errorf("invalid line detected: comma not allowed")
	}
	line = strings.Trim(line, `'"`)
	return line, nil
}
