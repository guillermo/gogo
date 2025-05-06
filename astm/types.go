package astm

import (
	"go/ast"
	"go/token"
)

// Code represents a Go source code file that can be modified
type Code struct {
	fset *token.FileSet
	file *ast.File
}
