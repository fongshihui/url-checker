package main

import (
	"bufio"
	"os"
	"strings"
)

func loadURLs(filePath string, args []string) ([]string, error) {
	if filePath == "" {
		return normalizeList(args), nil
	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []string
	scanner := bufio.NewScanner(f)
	// Skip blanks and comments.
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return normalizeList(out), nil
}

func normalizeList(in []string) []string {
	out := make([]string, 0, len(in))
	for _, raw := range in {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		out = append(out, raw)
	}
	return out
}
