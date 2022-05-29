package main

import (
	"fmt"
	"regexp"
)

func (d *Dependencies) formatPlaceholder(s string) (string, error) {
	r, err := regexp.Compile("$[0-9]+")
	if err != nil {
		return "", fmt.Errorf("failed to compile regexp: %w", err)
	}

	switch d.DatabaseProvider {
	case "sqlite3":
		fallthrough
	case "sqlite":
		fallthrough
	case "mysql":
		return r.ReplaceAllString(s, "?"), nil
	}

	return s, nil
}
