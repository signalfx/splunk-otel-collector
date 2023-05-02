package main

import (
	"errors"
	"strings"
	"unicode"
)

// See https://github.com/golang/lint/blob/d0100b6bd8b389f0385611eb39152c4d7c3a7905/lint.go#L771
var lintAcronyms = map[string]bool{
	"ACL":   true,
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XMPP":  true,
	"XSRF":  true,
	"XSS":   true,
}

// format variable in a go-safe way
func formatVariable(s string) (string, error) {
	if s == "" {
		return "", errors.New("string cannot be empty")
	}
	// Convert various characters to . for strings.Title to operate on.
	replace := strings.NewReplacer("_", ".", "-", ".", "<", ".", ">", ".", "/", ".", ":", ".")
	str := replace.Replace(s)
	str = strings.Title(str)
	str = strings.ReplaceAll(str, ".", "")

	var word string
	var output string

	// Fixup acronyms to make lint happy.
	for idx, r := range str {
		if idx == 0 {
			r = unicode.ToLower(r)
		}

		if unicode.IsUpper(r) || unicode.IsNumber(r) {
			if lintAcronyms[strings.ToUpper(word)] && output != "" {
				output += strings.ToUpper(word)
				word = string(r)
			} else {
				output += word
				word = string(r)
			}
		} else {
			word += string(r)
		}
	}

	if lintAcronyms[strings.ToUpper(word)] && output != "" {
		output += strings.ToUpper(word)
	} else {
		output += word
	}

	// Remove white spaces
	output = strings.Join(strings.Fields(output), "")

	return output, nil
}
