package astm

import (
	"go/parser"
	"go/token"
)

// Parse parses the Go source code and returns a Code object
func Parse(src []byte) *Code {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil
	}
	return &Code{fset: fset, file: file}
}
