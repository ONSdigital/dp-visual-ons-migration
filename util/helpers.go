package util

import (
	"regexp"
	"os"
	"errors"
	"strings"
)

var (
	validFilePattern   = "[^a-zA-Z0-9]+"
	validFileNameRegex *regexp.Regexp
)

func init() {
	var err error
	validFileNameRegex, err = regexp.Compile(validFilePattern)
	if err != nil {
		panic(errors.New("failed to compile valid filename regex"))
		os.Exit(1)
	}
}

func SanitisedFilename(name string) string {
	return strings.ToLower(validFileNameRegex.ReplaceAllString(name, ""))
}

