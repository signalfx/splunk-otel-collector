package selfdescribe

import (
	"go/ast"
	"go/doc"
	"regexp"
)

// Go's go/doc package has a note parser that collapses all whitespace to a
// single space instead of just letting the output renderer decide whether to
// keep the output or not.  Therefore, we have to copy/paste the note parsing
// code here and remove the whitespace normalization so that it is more useful.

var (
	noteMarker    = `([A-Z][A-Z_-]+)\(([^)]+)\):?`                  // MARKER(uid), MARKER at least 2 chars, uid at least 1 char
	noteMarkerRx  = regexp.MustCompile(`^[ \t]*` + noteMarker)      // MARKER(uid) at text start
	noteCommentRx = regexp.MustCompile(`^/[/*][ \t]*` + noteMarker) // MARKER(uid) at comment start
)

// readNote collects a single note from a sequence of comments.
func readNote(list []*ast.Comment, notes map[string][]*doc.Note) {
	text := (&ast.CommentGroup{List: list}).Text()
	if m := noteMarkerRx.FindStringSubmatchIndex(text); m != nil {
		// The note body starts after the marker.
		// We remove any formatting so that we don't
		// get spurious line breaks/indentation when
		// showing the TODO body.
		body := text[m[1]:]
		if body != "" {
			marker := text[m[2]:m[3]]
			notes[marker] = append(notes[marker], &doc.Note{
				Pos:  list[0].Pos(),
				End:  list[len(list)-1].End(),
				UID:  text[m[4]:m[5]],
				Body: body,
			})
		}
	}
}

// readNotes extracts notes from comments.
// A note must start at the beginning of a comment with "MARKER(uid):"
// and is followed by the note body (e.g., "// BUG(gri): fix this").
// The note ends at the end of the comment group or at the start of
// another note in the same comment group, whichever comes first.
func readNotes(comments []*ast.CommentGroup) map[string][]*doc.Note {
	notes := make(map[string][]*doc.Note)
	for _, group := range comments {
		i := -1 // comment index of most recent note start, valid if >= 0
		list := group.List
		for j, c := range list {
			if noteCommentRx.MatchString(c.Text) {
				if i >= 0 {
					readNote(list[i:j], notes)
				}
				i = j
			}
		}
		if i >= 0 {
			readNote(list[i:], notes)
		}
	}
	return notes
}
