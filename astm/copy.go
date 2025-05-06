package astm

import (
	"fmt"
	"go/ast"
	"go/token"
)

// Copy creates a copy of a declaration with a new name
func (c *Code) Copy(from, to string) error {
	// Find the declaration
	var decl interface{}
	ast.Inspect(c.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.Name == from {
				decl = x
				return false
			}
		case *ast.TypeSpec:
			if x.Name.Name == from {
				decl = x
				return false
			}
		case *ast.ValueSpec:
			for _, name := range x.Names {
				if name.Name == from {
					decl = x
					return false
				}
			}
		case *ast.Field:
			for _, name := range x.Names {
				if name.Name == from {
					decl = x
					return false
				}
			}
		}
		return true
	})

	if decl == nil {
		return fmt.Errorf("identifier %s not found", from)
	}

	// Create a new declaration with the new name
	switch d := decl.(type) {
	case *ast.FuncDecl:
		newDecl := &ast.FuncDecl{
			Recv: d.Recv,
			Name: &ast.Ident{Name: to},
			Type: d.Type,
			Body: d.Body,
		}
		c.file.Decls = append(c.file.Decls, newDecl)
	case *ast.TypeSpec:
		newSpec := &ast.TypeSpec{
			Name: &ast.Ident{Name: to},
			Type: d.Type,
		}
		newDecl := &ast.GenDecl{
			Tok:   token.TYPE,
			Specs: []ast.Spec{newSpec},
		}
		c.file.Decls = append(c.file.Decls, newDecl)
	case *ast.ValueSpec:
		newSpec := &ast.ValueSpec{
			Names:  []*ast.Ident{{Name: to}},
			Type:   d.Type,
			Values: d.Values,
		}
		newDecl := &ast.GenDecl{
			Tok:   token.VAR,
			Specs: []ast.Spec{newSpec},
		}
		c.file.Decls = append(c.file.Decls, newDecl)
	case *ast.Field:
		// If it's a field, find the containing struct
		var structType *ast.StructType
		ast.Inspect(c.file, func(n ast.Node) bool {
			if typeSpec, ok := n.(*ast.TypeSpec); ok {
				if structType, ok = typeSpec.Type.(*ast.StructType); ok {
					for _, field := range structType.Fields.List {
						if field == d {
							return false
						}
					}
				}
			}
			return true
		})

		if structType != nil {
			// Create a new field with the new name
			newField := &ast.Field{
				Names: []*ast.Ident{{Name: to}},
				Type:  d.Type,
				Tag:   d.Tag,
			}
			structType.Fields.List = append(structType.Fields.List, newField)
		}
	}

	return nil
}
