package astm

import (
	"go/ast"
)

// FieldRename renames a field in a struct
func (c *Code) FieldRename(structName, old, new string) error {
	// First, find the struct type declaration
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

	// Find and rename the field
	var fieldFound bool
	for _, field := range structType.Fields.List {
		for i, name := range field.Names {
			if name.Name == old {
				// Rename the field declaration
				field.Names[i].Name = new
				fieldFound = true
				break
			}
		}
	}

	if !fieldFound {
		return nil // Field not found
	}

	// Find and rename all references to this field
	ast.Inspect(c.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			// Handle field access (e.g., s.OldField)
			if x.Sel.Name == old {
				x.Sel.Name = new
			}
		case *ast.KeyValueExpr:
			// Handle struct literals (e.g., MyStruct{OldField: 1})
			if ident, ok := x.Key.(*ast.Ident); ok && ident.Name == old {
				ident.Name = new
			}
		}
		return true
	})

	return nil
}
