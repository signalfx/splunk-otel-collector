package utils

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode"
)

// FirstNonEmpty returns the first string that is not empty, otherwise ""
func FirstNonEmpty(s ...string) string {
	for _, str := range s {
		if str != "" {
			return str
		}
	}

	return ""
}

// FirstNonZero returns the first int in `ns` that is not zero.
func FirstNonZero(ns ...int) int {
	for _, n := range ns {
		if n != 0 {
			return n
		}
	}
	return 0
}

// IndentLines indents all lines in `ss` by `spaces` number of spaces
func IndentLines(ss string, spaces int) string {
	var output string
	for i := range ss {
		switch {
		case i == 0:
			output += strings.Repeat(" ", spaces) + string(ss[i])
		case ss[i] == '\n' && i != len(ss)-1:
			output += "\n" + strings.Repeat(" ", spaces)
		default:
			output += string(ss[i])
		}
	}
	return output
}

// LowercaseFirstChar make the first character of a string lowercase
func LowercaseFirstChar(s string) string {
	for i, v := range s {
		return string(unicode.ToLower(v)) + s[i+1:]
	}
	return ""
}

// ChunkScanner looks for a line and all subsequent indented lines and
// returns a scanner that will output that chunk as a single token.  This
// assumes that the entire chunk comes in a single read call, which will not
// always be the case.
func ChunkScanner(output io.Reader) *bufio.Scanner {
	s := bufio.NewScanner(output)
	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		lines := bytes.Split(data, []byte{'\n'})
		// If there is no newline in the data, lines will only have one element,
		// so return and wait for more data.
		if len(lines) == 1 && !atEOF {
			return 0, nil, nil
		}

		// For any subsequent indented lines, assume they are part of the same
		// log entry.  This requires that the whole entry be fed to this
		// function in a single chunk, so some entries may get split up
		// erroneously.
		var i int
		for i = 1; i < len(lines) && len(lines[i]) > 0 && (lines[i][0] == ' ' || lines[i][0] == '\t'); i++ {
		}

		entry := bytes.Join(lines[:i], []byte("\n"))
		// the above Join adds back all newlines lost except for one
		return len(entry) + 1, entry, nil
	})
	return s
}

func TrimAllSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, ch := range s {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func SplitString(s string, sep, escape rune) (tokens []string, err error) {
	var runes []rune
	inEscape := false
	for _, r := range s {
		switch {
		case inEscape:
			inEscape = false
			runes = append(runes, r)
		case r == escape:
			inEscape = true
		case r == sep:
			tokens = append(tokens, string(runes))
			runes = runes[:0]
		default:
			runes = append(runes, r)
		}
	}
	tokens = append(tokens, string(runes))
	if inEscape {
		err = errors.New("invalid terminal escape")
	}
	return tokens, err
}
