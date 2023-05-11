package selfdescribe

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type astCacheEntry struct {
	fset *token.FileSet
	pkgs map[string]*ast.Package
}

// nolint: gochecknoglobals
// Used to cache the asts parsed from package files
var astCache = map[string]astCacheEntry{}

// Returns the ast node of the struct itself and the comment group on the
// struct type.
func structNodes(packageDir, structName string) (*ast.TypeSpec, *ast.CommentGroup) {
	var fset *token.FileSet
	var pkgs map[string]*ast.Package

	cached, ok := astCache[packageDir]
	if ok {
		fset = cached.fset
		pkgs = cached.pkgs
	} else {
		fset = token.NewFileSet()
		var err error
		pkgs, err = parser.ParseDir(fset, packageDir, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		astCache[packageDir] = astCacheEntry{fset: fset, pkgs: pkgs}
	}

	for _, p := range pkgs {
		for _, f := range p.Files {
			// Find the struct specified by structName by looking at all nodes
			// with comments.  This means that the config struct has to have a
			// comment on it or else it won't be found.
			cmap := ast.NewCommentMap(fset, f, f.Comments)
			for node := range cmap {
				if t, ok := node.(*ast.GenDecl); ok {
					if t.Tok != token.TYPE {
						continue
					}

					if t.Specs[0].(*ast.TypeSpec).Name.Name == structName {
						return t.Specs[0].(*ast.TypeSpec), t.Doc
					}
				}
			}
		}
	}
	panic(fmt.Sprintf("Could not find %s in %s", structName, packageDir))
}

func structDoc(packageDir, structName string) string {
	_, commentGroup := structNodes(packageDir, structName)
	return commentTextToParagraphs(commentGroup.Text())
}

func packageDoc(packageDir string) *doc.Package {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, packageDir, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	if len(pkgs) > 1 {
		panic("Can't handle multiple packages")
	}
	if len(pkgs) == 0 {
		return nil
	}
	p := pkgs[filepath.Base(packageDir)]
	// go/doc is pretty inflexible in how it parses notes so do it ourselves.
	notes := readNotes(ast.MergePackageFiles(p, 0).Comments)
	pkgDoc := doc.New(p, packageDir, doc.AllDecls|doc.AllMethods)
	pkgDoc.Notes = notes
	return pkgDoc
}

func nestedPackageDocs(packageDir string) ([]*doc.Package, error) {
	var out []*doc.Package
	err := filepath.Walk(packageDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || err != nil {
			return err
		}

		pkgDoc := packageDoc(path)
		if pkgDoc != nil {
			out = append(out, pkgDoc)
		}
		return nil
	})
	return out, err
}

func notesFromDocs(docs []*doc.Package, noteType string) []*doc.Note {
	var notes []*doc.Note
	for _, pkgDoc := range docs {
		notes = append(notes, pkgDoc.Notes[noteType]...)
	}
	return notes
}

func structFieldDocs(packageDir, structName string) map[string]string {
	configStruct, _ := structNodes(packageDir, structName)
	fieldDocs := map[string]string{}
	for _, field := range configStruct.Type.(*ast.StructType).Fields.List {
		if field.Names != nil {
			fieldDocs[field.Names[0].Name] = commentTextToParagraphs(field.Doc.Text())
		}
	}

	return fieldDocs
}

// nolint: gochecknoglobals
var textRE = regexp.MustCompile(`([^\n])\n([^\s])`)

func commentTextToParagraphs(t string) string {
	return strings.TrimSpace(textRE.ReplaceAllString(t, "$1 $2"))
}
