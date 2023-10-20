package k8sutil

import (
	"regexp"
)

var re = regexp.MustCompile(`^[\w_-]+://`)

// StripContainerID returns a pure container id without the runtime scheme://
func StripContainerID(id string) string {
	return re.ReplaceAllString(id, "")
}
