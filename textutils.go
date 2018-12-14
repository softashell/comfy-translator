package main

import (
	"strings"
	"unicode"
)

func matchWhitespace(text string, source string) string {
	// TODO: Make this suck less

	text = strings.TrimSpace(text)
	text = extractLeadingWhitespace(source) + text + extractTrailingWhitespace(source)

	return text
}

func extractLeadingWhitespace(s string) string {
	var ws string

	for _, r := range s {
		if unicode.IsSpace(r) {
			ws += string(r)
			continue
		}

		break
	}

	return ws
}

func extractTrailingWhitespace(s string) string {
	var ws string

	// Unicode A SHIT
	reverse := func(s string) string {
		r := []rune(s)

		for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}

		return string(r)
	}

	s = reverse(s)

	ws = extractLeadingWhitespace(s)

	return reverse(ws)
}
