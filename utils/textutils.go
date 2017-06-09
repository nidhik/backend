package utils

import (
	"regexp"
	"strings"
)

func Capitalize(s string) string {
	r := regexp.MustCompile(`^.`)
	return r.ReplaceAllStringFunc(s, strings.ToUpper)
}
