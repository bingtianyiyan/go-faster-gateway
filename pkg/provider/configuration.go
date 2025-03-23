package provider

import (
	"strings"
	"unicode"
)

// Normalize replaces all special chars with `-`.
func Normalize(name string) string {
	fargs := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// get function
	return strings.Join(strings.FieldsFunc(name, fargs), "-")
}
