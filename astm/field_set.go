package astm

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// FieldSet sets a field in a struct
func (c *Code) FieldSet(structName, fieldName, fieldType, tags string) error {
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

	// Parse the field type
	expr, err := parser.ParseExpr(fieldType)
	if err != nil {
		return err
	}

	// Create the new field
	field := &ast.Field{
		Names: []*ast.Ident{{Name: fieldName}},
		Type:  expr,
	}

	// Add tags if provided
	if tags != "" {
		field.Tag = &ast.BasicLit{
			Kind:  token.STRING,
			Value: tags,
		}
	}

	// Add the field to the struct
	structType.Fields.List = append(structType.Fields.List, field)

	return nil
}
