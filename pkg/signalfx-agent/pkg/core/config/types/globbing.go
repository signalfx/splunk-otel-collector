package types

import (
	"strings"

	"github.com/gobwas/glob"
)

// PrefixAndGlob takes a path that is potentially globbed and returns the
// longest prefix of the path before a slash-delimited part with globbing
// characters.  E.g. if the path is "/a/b/c*/d", it would return "/a/b".  It
// also returns a glob.Glob instance that can be used to match the path.  The
// third return value tells whether the path has globs in it or not.  Some KV
// stores have no concept of directories (e.g. etcd3) and so they don't
// actually need to go back to the previous /, but watch logic should still
// work with them, albeit not quite as optimal.
func PrefixAndGlob(path string) (string, glob.Glob, bool, error) {
	prefix := path
	quoted := glob.QuoteMeta(path)
	for i, c := range quoted {
		if i >= len(path) || rune(path[i]) != c {
			prevSlashIdx := strings.LastIndex(quoted[:i], "/")
			if prevSlashIdx != -1 {
				prefix = quoted[:prevSlashIdx]
			}
			break
		}
	}
	g, err := glob.Compile(path)
	return prefix, g, quoted != path, err
}
