package template

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/guillermo/gogo"
)

// AddStructField adds a new field to a struct
func (t *Template) AddStructField(structName string, field gogo.StructField) (*Template, error) {
	// Check if the struct exists
	genDecl, typeSpec := t.findStructGenDecl(structName)
	if genDecl == nil || typeSpec == nil {
		return nil, fmt.Errorf("struct %s not found", structName)
	}

	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("%s is not a struct type", structName)
	}

	// Check if field already exists
	for _, existingField := range structType.Fields.List {
		for _, fieldName := range existingField.Names {
			if fieldName.Name == field.Name {
				return nil, fmt.Errorf("field %s already exists in struct %s", field.Name, structName)
			}
		}
	}

	// Create a new template with the changes
	newTemplate := t.cloneTemplate()

	// Find the struct in the new template
	newGenDecl, newTypeSpec := newTemplate.findStructGenDecl(structName)
	if newGenDecl == nil || newTypeSpec == nil {
		return nil, fmt.Errorf("failed to find struct in cloned template")
	}

	newStructType, ok := newTypeSpec.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("failed to get struct type in cloned template")
	}

	// Create the new field
	newField := &ast.Field{
		Names: []*ast.Ident{{Name: field.Name}},
		Type:  parseTypeExpr(field.Type),
	}

	// Add tag/annotation if present
	if field.Annotation != "" {
		// Ensure the annotation has backticks
		annotation := field.Annotation
		if len(annotation) > 0 && annotation[0] != '`' {
			annotation = "`" + annotation + "`"
		}
		newField.Tag = &ast.BasicLit{
			Kind:  token.STRING,
			Value: annotation,
		}
	}

	// Add the field to the struct
	newStructType.Fields.List = append(newStructType.Fields.List, newField)

	return newTemplate, nil
}

// RemoveStructField removes a field from a struct
func (t *Template) RemoveStructField(structName string, field gogo.StructField) (*Template, error) {
	// Check if the struct exists
	genDecl, typeSpec := t.findStructGenDecl(structName)
	if genDecl == nil || typeSpec == nil {
		return nil, fmt.Errorf("struct %s not found", structName)
	}

	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("%s is not a struct type", structName)
	}

	// Check if field exists
	fieldFound := false
	for _, existingField := range structType.Fields.List {
		for _, fieldName := range existingField.Names {
			if fieldName.Name == field.Name {
				fieldFound = true
				break
			}
		}
		if fieldFound {
			break
		}
	}

	if !fieldFound {
		return nil, fmt.Errorf("field %s not found in struct %s", field.Name, structName)
	}

	// Create a new template with the changes
	newTemplate := t.cloneTemplate()

	// Find the struct in the new template
	newGenDecl, newTypeSpec := newTemplate.findStructGenDecl(structName)
	if newGenDecl == nil || newTypeSpec == nil {
		return nil, fmt.Errorf("failed to find struct in cloned template")
	}

	newStructType, ok := newTypeSpec.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("failed to get struct type in cloned template")
	}

	// Remove the field from the struct
	newFields := make([]*ast.Field, 0, len(newStructType.Fields.List))
	for _, existingField := range newStructType.Fields.List {
		shouldRemove := false
		for _, fieldName := range existingField.Names {
			if fieldName.Name == field.Name {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			newFields = append(newFields, existingField)
		}
	}

	newStructType.Fields.List = newFields

	return newTemplate, nil
}
