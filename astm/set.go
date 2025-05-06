package astm

import (
	"go/parser"
	"go/token"
)

// Set adds new code to the tree
func (c *Code) Set(src []byte) error {
	// Parse the new code
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return err
	}

	// Add all declarations from the new code to the existing file
	for _, decl := range file.Decls {
		c.file.Decls = append(c.file.Decls, decl)
	}

	return nil
}
