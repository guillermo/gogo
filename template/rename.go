package template

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/guillermo/gogo"
)

// RenameStruct renames a struct and all references to it throughout the AST
func (t *Template) RenameStruct(oldName, newName string) (*Template, error) {
	// Check if the struct exists
	if t.findStruct(oldName) == nil {
		return nil, fmt.Errorf("struct %s not found", oldName)
	}

	// Create a new template with the changes
	newTemplate := t.cloneTemplate()

	// Rename the struct declaration and all references
	newTemplate.renameIdentifier(oldName, newName)

	return newTemplate, nil
}

// RenameStructField renames a struct field and all references to it throughout the AST.
// It can also update the field's type and annotation if specified in newField.
func (t *Template) RenameStructField(structName, oldFieldName string, newField gogo.StructField) (*Template, error) {
	// Check if the struct exists
	structType := t.findStruct(structName)
	if structType == nil {
		return nil, fmt.Errorf("struct %s not found", structName)
	}

	// Check if the old field exists in the struct
	fieldFound := false
	for _, field := range structType.Fields.List {
		for _, fieldName := range field.Names {
			if fieldName.Name == oldFieldName {
				fieldFound = true
				break
			}
		}
		if fieldFound {
			break
		}
	}

	if !fieldFound {
		return nil, fmt.Errorf("field %s not found in struct %s", oldFieldName, structName)
	}

	// Create a new template with the changes
	newTemplate := t.cloneTemplate()

	// Rename the field in the struct definition and all selector expressions
	newTemplate.renameStructFieldIdentifier(structName, oldFieldName, newField.Name)

	// Update the field type and annotation if specified
	genDecl, typeSpec := newTemplate.findStructGenDecl(structName)
	if genDecl != nil && typeSpec != nil {
		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
			for _, field := range structType.Fields.List {
				for _, fieldName := range field.Names {
					if fieldName.Name == newField.Name {
						// Update type if specified
						if newField.Type != "" {
							field.Type = parseTypeExpr(newField.Type)
						}

						// Update annotation if specified
						if newField.Annotation != "" {
							annotation := newField.Annotation
							if len(annotation) > 0 && annotation[0] != '`' {
								annotation = "`" + annotation + "`"
							}
							field.Tag = &ast.BasicLit{
								Kind:  token.STRING,
								Value: annotation,
							}
						}
					}
				}
			}
		}
	}

	return newTemplate, nil
}

// RenameVariable renames a variable and all references to it throughout the AST
func (t *Template) RenameVariable(oldName, newName string) (*Template, error) {
	// Check if the variable exists
	if t.findVariable(oldName) == nil {
		return nil, fmt.Errorf("variable %s not found", oldName)
	}

	// Create a new template with the changes
	newTemplate := t.cloneTemplate()

	// Rename the variable declaration and all references
	newTemplate.renameIdentifier(oldName, newName)

	return newTemplate, nil
}

// RenameFunction renames a function and all references to it throughout the AST
func (t *Template) RenameFunction(oldName, newName string) (*Template, error) {
	// Check if the function exists
	if t.findFunction(oldName) == nil {
		return nil, fmt.Errorf("function %s not found", oldName)
	}

	// Create a new template with the changes
	newTemplate := t.cloneTemplate()

	// Rename the function declaration and all references
	newTemplate.renameIdentifier(oldName, newName)

	return newTemplate, nil
}

// RenameType renames a type and all references to it throughout the AST
func (t *Template) RenameType(oldName, newName string) (*Template, error) {
	// Check if the type exists
	if t.findType(oldName) == nil {
		return nil, fmt.Errorf("type %s not found", oldName)
	}

	// Create a new template with the changes
	newTemplate := t.cloneTemplate()

	// Rename the type declaration and all references
	newTemplate.renameIdentifier(oldName, newName)

	return newTemplate, nil
}

// RenameConstant renames a constant and all references to it throughout the AST
func (t *Template) RenameConstant(oldName, newName string) (*Template, error) {
	// Check if the constant exists
	if t.findConstant(oldName) == nil {
		return nil, fmt.Errorf("constant %s not found", oldName)
	}

	// Create a new template with the changes
	newTemplate := t.cloneTemplate()

	// Rename the constant declaration and all references
	newTemplate.renameIdentifier(oldName, newName)

	return newTemplate, nil
}

// parseTypeExpr is imported from parser.go for consistency
// It parses a type string into an AST expression
func parseTypeExpr(typeStr string) ast.Expr {
	// Handle basic types
	if !contains(typeStr, ".") && !contains(typeStr, "[") && !contains(typeStr, "*") {
		return &ast.Ident{Name: typeStr}
	}

	// Handle pointer types
	if len(typeStr) > 0 && typeStr[0] == '*' {
		return &ast.StarExpr{
			X: parseTypeExpr(typeStr[1:]),
		}
	}

	// Handle slice types
	if len(typeStr) > 2 && typeStr[0:2] == "[]" {
		return &ast.ArrayType{
			Elt: parseTypeExpr(typeStr[2:]),
		}
	}

	// Handle map types
	if len(typeStr) > 4 && typeStr[0:4] == "map[" {
		// Simple map parsing - could be enhanced
		closeIdx := indexOf(typeStr, "]")
		if closeIdx > 0 {
			keyType := typeStr[4:closeIdx]
			valueType := typeStr[closeIdx+1:]
			return &ast.MapType{
				Key:   parseTypeExpr(keyType),
				Value: parseTypeExpr(valueType),
			}
		}
	}

	// Handle qualified types (package.Type)
	if idx := lastIndexOf(typeStr, "."); idx > 0 {
		return &ast.SelectorExpr{
			X:   &ast.Ident{Name: typeStr[:idx]},
			Sel: &ast.Ident{Name: typeStr[idx+1:]},
		}
	}

	// Default to identifier
	return &ast.Ident{Name: typeStr}
}

// Helper functions
func contains(s, substr string) bool {
	return indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func lastIndexOf(s, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
