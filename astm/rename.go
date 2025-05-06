package astm

import (
	"go/ast"
)

// Rename renames an identifier and its references
func (c *Code) Rename(old, new string) error {
	// First, find all declarations of the identifier
	var decls []ast.Node
	ast.Inspect(c.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if x.Name.Name == old {
				decls = append(decls, x)
			}
		case *ast.FuncDecl:
			if x.Name.Name == old {
				decls = append(decls, x)
			}
		case *ast.ValueSpec:
			for _, name := range x.Names {
				if name.Name == old {
					decls = append(decls, x)
					break
				}
			}
		}
		return true
	})

	if len(decls) == 0 {
		return nil // No declarations found, nothing to rename
	}

	// Then, find and rename all references
	ast.Inspect(c.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if x.Name == old {
				x.Name = new
			}
		}
		return true
	})

	// Finally, rename the declarations
	for _, decl := range decls {
		switch x := decl.(type) {
		case *ast.TypeSpec:
			x.Name.Name = new
		case *ast.FuncDecl:
			x.Name.Name = new
		case *ast.ValueSpec:
			for _, name := range x.Names {
				if name.Name == old {
					name.Name = new
				}
			}
		}
	}

	return nil
}
