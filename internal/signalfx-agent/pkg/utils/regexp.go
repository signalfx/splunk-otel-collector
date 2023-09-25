package utils

import (
	"regexp"
)

// RegexpGroupMap matches text against the given regexp and returns a map of
// all of the named subgroups to the values found in text.  Returns nil if text
// does not match.
func RegexpGroupMap(re *regexp.Regexp, text string) map[string]string {
	if groups := re.FindStringSubmatch(text); groups != nil {
		groupMap := map[string]string{}
		for i, name := range re.SubexpNames() {
			groupMap[name] = groups[i]
		}
		return groupMap
	}
	return nil
}
