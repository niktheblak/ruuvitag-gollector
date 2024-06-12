package psql

import (
	"regexp"
)

var (
	passwordRegexp   = regexp.MustCompile(`password=\S+\s`)
	whitespaceRegexp = regexp.MustCompile(`\s+`)
)

// RemovePassword removes password from a psqlInfo style string
func RemovePassword(psqlInfo string) string {
	return passwordRegexp.ReplaceAllString(psqlInfo, "password=[redacted] ")
}

// TrimQuery replaces all whitespace (newlines and repeated spaces) from a string with one space
func TrimQuery(q string) string {
	return whitespaceRegexp.ReplaceAllString(q, " ")
}
