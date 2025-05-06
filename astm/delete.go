package astm

import (
	"fmt"
	"go/ast"
)

// Delete removes an identifier and its references
func (c *Code) Delete(name string) error {
	// First, find all references to the identifier
	var refs []*ast.Ident
	ast.Inspect(c.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if x.Name == name {
				refs = append(refs, x)
			}
		}
		return true
	})

	// Then, find the declaration
	var decl interface{}
	var declIdent *ast.Ident
	for _, ref := range refs {
		if ref.Obj != nil && ref.Obj.Decl != nil {
			switch d := ref.Obj.Decl.(type) {
			case *ast.FuncDecl:
				if d.Name == ref {
					decl = d
					declIdent = ref
					break
				}
			case *ast.TypeSpec:
				if d.Name == ref {
					decl = d
					declIdent = ref
					break
				}
			case *ast.ValueSpec:
				for _, n := range d.Names {
					if n == ref {
						decl = d
						declIdent = ref
						break
					}
				}
			}
		}
	}

	if decl == nil {
		return nil // Nothing to delete
	}

	// Check if there are any references other than the declaration
	var hasRefs bool
	for _, ref := range refs {
		if ref == declIdent {
			continue // Skip the declaration
		}
		hasRefs = true
		break
	}

	if hasRefs {
		return fmt.Errorf("cannot delete %s: it has references", name)
	}

	// Finally, remove the declaration
	var newDecls []ast.Decl
	for _, d := range c.file.Decls {
		keep := true
		switch x := d.(type) {
		case *ast.FuncDecl:
			if x == decl {
				keep = false
			}
		case *ast.GenDecl:
			var newSpecs []ast.Spec
			for _, spec := range x.Specs {
				keepSpec := true
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s == decl {
						keepSpec = false
					}
				case *ast.ValueSpec:
					if s == decl {
						keepSpec = false
					}
				}
				if keepSpec {
					newSpecs = append(newSpecs, spec)
				}
			}
			if len(newSpecs) == 0 {
				keep = false
			} else {
				x.Specs = newSpecs
			}
		}
		if keep {
			newDecls = append(newDecls, d)
		}
	}
	c.file.Decls = newDecls

	return nil
}
