package astm

import (
	"go/ast"
)

// FieldDelete deletes a field from a struct
func (c *Code) FieldDelete(structName, fieldName string) error {
	// Find the struct type declaration
	var structType *ast.StructType
	ast.Inspect(c.file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if typeSpec.Name.Name == structName {
				if structType, ok = typeSpec.Type.(*ast.StructType); ok {
					return false // Found the struct, stop searching
				}
			}
		}
		return true
	})

	if structType == nil {
		return nil // Struct not found
	}

	// Find and remove the field
	var newFields []*ast.Field
	for _, field := range structType.Fields.List {
		keep := true
		for _, name := range field.Names {
			if name.Name == fieldName {
				keep = false
				break
			}
		}
		if keep {
			newFields = append(newFields, field)
		}
	}
	structType.Fields.List = newFields

	// Remove references to the field in struct literals
	ast.Inspect(c.file, func(n ast.Node) bool {
		if compositeLit, ok := n.(*ast.CompositeLit); ok {
			if ident, ok := compositeLit.Type.(*ast.Ident); ok && ident.Name == structName {
				var newElts []ast.Expr
				for _, elt := range compositeLit.Elts {
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name != fieldName {
							newElts = append(newElts, elt)
						}
					} else {
						newElts = append(newElts, elt)
					}
				}
				compositeLit.Elts = newElts
			}
		}
		return true
	})

	return nil
}
