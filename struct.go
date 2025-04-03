package gogo

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// StructModifier is a function type that modifies a struct.
// It's passed to methods like OpenStruct to apply changes to a struct.
type StructModifier func(s Struct)

// Field represents a struct field with its name, type, and tags.
// It provides a more convenient interface than the ast.Field type.
type Field struct {
	Name     string            // Name of the field
	Type     string            // Type of the field
	Tags     map[string]string // Map of field tags
	AstField *ast.Field        // Reference to the original AST field
}

// SetType changes the type of the field and updates the underlying AST.
func (f *Field) SetType(typeName string) {
	f.Type = typeName
	f.AstField.Type = ast.NewIdent(typeName)
}

// SetTags sets the tags of the field and updates the underlying AST.
func (f *Field) SetTags(tags map[string]string) {
	f.Tags = tags

	// Create the tag string
	if len(tags) > 0 {
		tagParts := make([]string, 0, len(tags))
		for k, v := range tags {
			tagParts = append(tagParts, fmt.Sprintf("%s:\"%s\"", k, v))
		}
		tagStr := strings.Join(tagParts, " ")

		f.AstField.Tag = &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`" + tagStr + "`",
		}
	} else {
		f.AstField.Tag = nil
	}
}

// AddTag adds or updates a single tag.
func (f *Field) AddTag(key, value string) {
	if f.Tags == nil {
		f.Tags = make(map[string]string)
	}
	f.Tags[key] = value
	f.SetTags(f.Tags)
}

// String returns a string representation of the field.
func (f *Field) String() string {
	// Create the tag string
	tagStr := ""
	if len(f.Tags) > 0 {
		tagParts := make([]string, 0, len(f.Tags))
		for k, v := range f.Tags {
			tagParts = append(tagParts, fmt.Sprintf("%s:\"%s\"", k, v))
		}
		tagStr = " `" + strings.Join(tagParts, " ") + "`"
	}

	return fmt.Sprintf("%s %s%s", f.Name, f.Type, tagStr)
}

// Struct represents a Go struct that can be manipulated.
// It provides methods to add, remove, rename, and modify struct fields and methods.
type Struct struct {
	Name        string          // Name is the name of the struct
	Fields      []*ast.Field    // Fields is the list of fields in the struct
	StructType  *ast.StructType // StructType is the AST representation of the struct
	Parent      *File           // Parent is the file that contains the struct
	Declaration *ast.TypeSpec   // Declaration is the type specification of the struct
}

// Field finds a field by name and returns a Field representation.
// Returns nil if the field is not found.
func (s Struct) Field(name string) *Field {
	for _, field := range s.Fields {
		if len(field.Names) > 0 && field.Names[0].Name == name {
			// Create a new Field object
			f := &Field{
				Name:     name,
				AstField: field,
			}

			// Extract type
			if typeIdent, ok := field.Type.(*ast.Ident); ok {
				f.Type = typeIdent.Name
			} else {
				// For complex types, use a simple string representation
				f.Type = fmt.Sprintf("%T", field.Type)
			}

			// Extract tags if present
			if field.Tag != nil {
				f.Tags = parseTags(field.Tag.Value)
			} else {
				f.Tags = make(map[string]string)
			}

			return f
		}
	}
	return nil // Return nil if field not found
}

// parseTags extracts field tags from a tag string
func parseTags(tagStr string) map[string]string {
	result := make(map[string]string)

	// Remove the backticks
	if len(tagStr) >= 2 {
		tagStr = tagStr[1 : len(tagStr)-1]
	}

	// Split by spaces
	parts := strings.Fields(tagStr)
	for _, part := range parts {
		// Split by colon
		keyValue := strings.SplitN(part, ":", 2)
		if len(keyValue) == 2 {
			// Remove quotes from value
			value := keyValue[1]
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			result[keyValue[0]] = value
		}
	}

	return result
}

// Add adds a field to a struct with the specified name, type, and tags.
// The tags parameter is a map of tag keys to values.
func (s Struct) Add(name, fieldType string, tags map[string]string) {
	// Create the tag string
	tagStr := ""
	if len(tags) > 0 {
		tagParts := make([]string, 0, len(tags))
		for k, v := range tags {
			tagParts = append(tagParts, fmt.Sprintf("%s:\"%s\"", k, v))
		}
		tagStr = strings.Join(tagParts, " ")
	}

	// Create the field
	field := &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(name)},
		Type:  ast.NewIdent(fieldType),
	}

	// Add the tag if provided
	if tagStr != "" {
		field.Tag = &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`" + tagStr + "`",
		}
	}

	// Add the field to the struct
	s.StructType.Fields.List = append(s.StructType.Fields.List, field)
}

// Remove removes a field from a struct by name.
// It returns an error if the field is not found.
func (s Struct) Remove(name string) error {
	found := false
	var newFields []*ast.Field

	for _, field := range s.Fields {
		shouldKeep := true
		for _, ident := range field.Names {
			if ident.Name == name {
				found = true
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			newFields = append(newFields, field)
		}
	}

	s.StructType.Fields.List = newFields
	if !found {
		return ErrNotFound
	}
	return nil
}

// RemoveMethod removes a method from the struct by name.
// If the method is called elsewhere in the code, it creates a stub implementation.
// It returns an error if the method is not found.
func (s Struct) RemoveMethod(methodName string) error {
	found := false
	file := s.Parent
	var removedMethod *ast.FuncDecl

	// First, find the method to save its signature
	for _, decl := range file.AstFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok &&
			funcDecl.Name.Name == methodName &&
			funcDecl.Recv != nil &&
			len(funcDecl.Recv.List) > 0 {

			// Check if the receiver type matches the struct name
			if receiverType, ok := getReceiverType(funcDecl); ok && receiverType == s.Name {
				found = true
				removedMethod = funcDecl
				break
			}
		}
	}

	if !found {
		return ErrNotFound
	}

	// Now, check if the method is called somewhere in the code
	methodCalled := false
	ast.Inspect(file.AstFile, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if selExpr.Sel.Name == methodName {
					methodCalled = true
				}
			}
		}
		return true
	})

	// If the method is called, we should create a stub implementation
	if methodCalled && removedMethod != nil {
		// Create a new method with a default implementation
		newMethod := &ast.FuncDecl{
			Name: ast.NewIdent(methodName),
			Type: removedMethod.Type,
			Recv: removedMethod.Recv,
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							// Return a default value based on return type
							getDefaultReturnValue(removedMethod.Type.Results),
						},
					},
				},
			},
		}

		// Add the new method to the file
		file.AstFile.Decls = append(file.AstFile.Decls, newMethod)
	} else {
		// Create a new slice of declarations excluding the method to remove
		var newDecls []ast.Decl
		for _, decl := range file.AstFile.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok &&
				funcDecl.Name.Name == methodName &&
				funcDecl.Recv != nil &&
				len(funcDecl.Recv.List) > 0 {

				// Check if the receiver type matches the struct name
				if receiverType, ok := getReceiverType(funcDecl); ok && receiverType == s.Name {
					// Skip this declaration
					continue
				}
			}
			newDecls = append(newDecls, decl)
		}

		// Update the AST with the new declarations
		file.AstFile.Decls = newDecls
	}

	return nil
}

// getDefaultReturnValue creates a default return value expression based on the return type.
// It handles primitive types like strings, integers, floats, and booleans, as well as complex types.
func getDefaultReturnValue(results *ast.FieldList) ast.Expr {
	if results == nil || len(results.List) == 0 {
		return nil
	}

	// Get the first return type
	retType := results.List[0].Type

	// Create default return values based on the type
	switch t := retType.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return &ast.BasicLit{
				Kind:  token.STRING,
				Value: `"HELLO, WORLD!"`,
			}
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			return &ast.BasicLit{
				Kind:  token.INT,
				Value: "0",
			}
		case "float32", "float64":
			return &ast.BasicLit{
				Kind:  token.FLOAT,
				Value: "0.0",
			}
		case "bool":
			return &ast.Ident{Name: "false"}
		default:
			return &ast.Ident{Name: "nil"}
		}
	case *ast.StarExpr, *ast.ArrayType, *ast.MapType, *ast.ChanType, *ast.InterfaceType:
		return &ast.Ident{Name: "nil"}
	default:
		return &ast.Ident{Name: "nil"}
	}
}

// DuplicateMethod duplicates a method with a new name
func (s Struct) DuplicateMethod(oldName, newName string) error {
	found := false
	file := s.Parent

	// Find the method to duplicate
	for _, decl := range file.AstFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok &&
			funcDecl.Name.Name == oldName &&
			funcDecl.Recv != nil &&
			len(funcDecl.Recv.List) > 0 {

			// Check if the receiver type matches the struct name
			if receiverType, ok := getReceiverType(funcDecl); ok && receiverType == s.Name {
				found = true

				// Create a deep copy of the method
				dupMethod := copyDecl(funcDecl).(*ast.FuncDecl)
				dupMethod.Name.Name = newName

				// Add the duplicated method to the file
				file.AstFile.Decls = append(file.AstFile.Decls, dupMethod)
				break
			}
		}
	}

	if !found {
		return ErrNotFound
	}
	return nil
}

// RenameMethod renames a method of the struct
func (s Struct) RenameMethod(oldName, newName string) error {
	found := false
	file := s.Parent

	// Find the method declaration and rename it
	for _, decl := range file.AstFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok &&
			funcDecl.Name.Name == oldName &&
			funcDecl.Recv != nil &&
			len(funcDecl.Recv.List) > 0 {

			// Check if the receiver type matches the struct name
			if receiverType, ok := getReceiverType(funcDecl); ok && receiverType == s.Name {
				funcDecl.Name.Name = newName
				found = true

				// Also update method calls in the file
				ast.Inspect(file.AstFile, func(n ast.Node) bool {
					if call, ok := n.(*ast.SelectorExpr); ok {
						if call.Sel.Name == oldName {
							// Check if this is a method call on our struct type
							call.Sel.Name = newName
						}
					}
					return true
				})

				break
			}
		}
	}

	if !found {
		return ErrNotFound
	}
	return nil
}

// getReceiverType extracts the receiver type name from a method declaration
func getReceiverType(funcDecl *ast.FuncDecl) (string, bool) {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return "", false
	}

	receiverType := funcDecl.Recv.List[0].Type

	// Handle pointer receiver: *Type
	if starExpr, ok := receiverType.(*ast.StarExpr); ok {
		if ident, ok := starExpr.X.(*ast.Ident); ok {
			return ident.Name, true
		}
	}

	// Handle non-pointer receiver: Type
	if ident, ok := receiverType.(*ast.Ident); ok {
		return ident.Name, true
	}

	return "", false
}

// AddField adds a field to the struct using a Field object.
// This is a more flexible alternative to the Add method.
func (s Struct) AddField(field *Field) {
	// Create a new ast.Field from the Field object
	astField := &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(field.Name)},
		Type:  ast.NewIdent(field.Type),
	}

	// Add tags if present
	if len(field.Tags) > 0 {
		tagParts := make([]string, 0, len(field.Tags))
		for k, v := range field.Tags {
			tagParts = append(tagParts, fmt.Sprintf("%s:\"%s\"", k, v))
		}
		tagStr := strings.Join(tagParts, " ")

		astField.Tag = &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`" + tagStr + "`",
		}
	}

	// Add the field to the struct
	s.StructType.Fields.List = append(s.StructType.Fields.List, astField)
}

// NewField creates a new Field object that can be used with AddField.
// This is useful when you want to create a field before adding it to a struct.
func NewField(name, fieldType string, tags map[string]string) *Field {
	field := &Field{
		Name: name,
		Type: fieldType,
		Tags: tags,
	}
	return field
}
