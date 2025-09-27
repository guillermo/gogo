package gogo

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
)

// parseFieldsString parses a field string like `ID uuid.UUID \`json:"id"\“ into StructField slice
func parseFieldsString(fields string) ([]StructField, error) {
	// Simple parser for field strings
	// This is a basic implementation - could be enhanced
	var result []StructField

	lines := strings.Split(fields, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse each line: Name Type `annotation`
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		field := StructField{
			Name: parts[0],
			Type: parts[1],
		}

		// Find annotation (everything after backtick)
		if idx := strings.Index(line, "`"); idx >= 0 {
			field.Annotation = line[idx:]
		}

		result = append(result, field)
	}

	return result, nil
}

// createNewFileWithStruct creates a new Go file with the specified struct
func createNewFileWithStruct(s structDef, packageName string) ([]byte, error) {
	if packageName == "" {
		packageName = "main"
	}

	var buf bytes.Buffer

	// Write package declaration
	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// Write struct
	buf.WriteString(fmt.Sprintf("type %s struct {\n", s.Name))

	for _, field := range s.EnsureFields {
		buf.WriteString(fmt.Sprintf("\t%s %s", field.Name, field.Type))
		if field.Annotation != "" {
			// Ensure the annotation has backticks if not already present
			annotation := field.Annotation
			if !strings.HasPrefix(annotation, "`") {
				annotation = "`" + annotation + "`"
			}
			buf.WriteString(" " + annotation)
		}
		buf.WriteString("\n")
	}

	buf.WriteString("}\n")

	// Format the code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), nil // Return unformatted if format fails
	}

	return formatted, nil
}

// modifyExistingFile modifies an existing Go file to ensure/delete struct fields
func modifyExistingFile(content []byte, s structDef) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	// Find or create the struct
	structType := findOrCreateStruct(file, s.Name)
	if structType == nil {
		// Add new struct to file
		structType = createStructDecl(s)
		file.Decls = append(file.Decls, structType)
	} else {
		// Modify existing struct
		modifyStruct(structType, s)
	}

	// Format and return the modified code
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("failed to format Go code: %w", err)
	}

	return buf.Bytes(), nil
}

// findOrCreateStruct finds an existing struct or returns nil if not found
func findOrCreateStruct(file *ast.File, name string) *ast.GenDecl {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if typeSpec.Name.Name == name {
				if _, ok := typeSpec.Type.(*ast.StructType); ok {
					return genDecl
				}
			}
		}
	}

	return nil
}

// createStructDecl creates a new struct declaration
func createStructDecl(s structDef) *ast.GenDecl {
	fields := &ast.FieldList{
		List: make([]*ast.Field, 0, len(s.EnsureFields)),
	}

	for _, field := range s.EnsureFields {
		astField := &ast.Field{
			Names: []*ast.Ident{{Name: field.Name}},
			Type:  parseTypeExpr(field.Type),
		}

		if field.Annotation != "" {
			// Ensure the annotation has backticks
			annotation := field.Annotation
			if !strings.HasPrefix(annotation, "`") {
				annotation = "`" + annotation + "`"
			}
			astField.Tag = &ast.BasicLit{
				Kind:  token.STRING,
				Value: annotation,
			}
		}

		fields.List = append(fields.List, astField)
	}

	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: s.Name},
				Type: &ast.StructType{
					Fields: fields,
				},
			},
		},
	}
}

// modifyStruct modifies an existing struct according to the Struct specification
func modifyStruct(genDecl *ast.GenDecl, s structDef) {
	for _, spec := range genDecl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok || typeSpec.Name.Name != s.Name {
			continue
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		// Build map of existing fields
		existingFields := make(map[string]*ast.Field)
		for _, field := range structType.Fields.List {
			for _, name := range field.Names {
				existingFields[name.Name] = field
			}
		}

		// Remove fields marked for deletion
		if len(s.DeleteFields) > 0 {
			newFields := make([]*ast.Field, 0)

			for _, field := range structType.Fields.List {
				shouldDelete := false

				// Check if field should be deleted
				for _, name := range field.Names {
					for _, deleteField := range s.DeleteFields {
						if name.Name == deleteField.Name {
							shouldDelete = true
							break
						}
					}
					if shouldDelete {
						break
					}
				}

				if !shouldDelete {
					newFields = append(newFields, field)
				}
			}

			structType.Fields.List = newFields

			// Rebuild existing fields map after deletion
			existingFields = make(map[string]*ast.Field)
			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					existingFields[name.Name] = field
				}
			}
		} else if !s.PreserveExisting {
			// If not preserving existing fields and not deleting specific ones,
			// clear all fields
			structType.Fields.List = make([]*ast.Field, 0)
			existingFields = make(map[string]*ast.Field)
		}

		// Add or update ensure fields
		for _, ensureField := range s.EnsureFields {
			if existing, ok := existingFields[ensureField.Name]; ok {
				// Update existing field
				existing.Type = parseTypeExpr(ensureField.Type)
				if ensureField.Annotation != "" {
					// Ensure the annotation has backticks
					annotation := ensureField.Annotation
					if !strings.HasPrefix(annotation, "`") {
						annotation = "`" + annotation + "`"
					}
					existing.Tag = &ast.BasicLit{
						Kind:  token.STRING,
						Value: annotation,
					}
				}
			} else {
				// Add new field
				newField := &ast.Field{
					Names: []*ast.Ident{{Name: ensureField.Name}},
					Type:  parseTypeExpr(ensureField.Type),
				}

				if ensureField.Annotation != "" {
					// Ensure the annotation has backticks
					annotation := ensureField.Annotation
					if !strings.HasPrefix(annotation, "`") {
						annotation = "`" + annotation + "`"
					}
					newField.Tag = &ast.BasicLit{
						Kind:  token.STRING,
						Value: annotation,
					}
				}

				structType.Fields.List = append(structType.Fields.List, newField)
			}
		}
	}
}

// parseTypeExpr parses a type string into an AST expression
func parseTypeExpr(typeStr string) ast.Expr {
	// Handle basic types
	if !strings.Contains(typeStr, ".") && !strings.Contains(typeStr, "[") && !strings.Contains(typeStr, "*") {
		return &ast.Ident{Name: typeStr}
	}

	// Handle pointer types
	if strings.HasPrefix(typeStr, "*") {
		return &ast.StarExpr{
			X: parseTypeExpr(typeStr[1:]),
		}
	}

	// Handle slice types
	if strings.HasPrefix(typeStr, "[]") {
		return &ast.ArrayType{
			Elt: parseTypeExpr(typeStr[2:]),
		}
	}

	// Handle map types
	if strings.HasPrefix(typeStr, "map[") {
		// Simple map parsing - could be enhanced
		closeIdx := strings.Index(typeStr, "]")
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
	if idx := strings.LastIndex(typeStr, "."); idx > 0 {
		return &ast.SelectorExpr{
			X:   &ast.Ident{Name: typeStr[:idx]},
			Sel: &ast.Ident{Name: typeStr[idx+1:]},
		}
	}

	// Default to identifier
	return &ast.Ident{Name: typeStr}
}

// createNewFileWithMethod creates a new Go file with the specified method
func createNewFileWithMethod(opts MethodOpts, packageName string) ([]byte, error) {
	if packageName == "" {
		packageName = "main"
	}

	var buf bytes.Buffer

	// Write package declaration
	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// If Content is provided, use it directly
	if opts.Content != "" {
		// Determine receiver name
		receiverName := opts.ReceiverName
		if receiverName == "" {
			// Default receiver name (first letter of type, skipping pointer)
			typeWithoutPointer := strings.TrimPrefix(opts.ReceiverType, "*")
			if len(typeWithoutPointer) > 0 {
				receiverName = strings.ToLower(string(typeWithoutPointer[0]))
			} else {
				receiverName = "r"
			}
		}
		buf.WriteString(fmt.Sprintf("func (%s %s) %s%s\n", receiverName, opts.ReceiverType, opts.Name, opts.Content))
	} else {
		// Write method
		receiverName := opts.ReceiverName
		if receiverName == "" {
			// Default receiver name (first letter of type, skipping pointer)
			typeWithoutPointer := strings.TrimPrefix(opts.ReceiverType, "*")
			if len(typeWithoutPointer) > 0 {
				receiverName = strings.ToLower(string(typeWithoutPointer[0]))
			} else {
				receiverName = "r"
			}
		}

		buf.WriteString(fmt.Sprintf("func (%s %s) %s(", receiverName, opts.ReceiverType, opts.Name))

		// Add parameters
		for i, param := range opts.Parameters {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(fmt.Sprintf("%s %s", param.Name, param.Type))
		}

		buf.WriteString(")")

		// Add return type
		if opts.ReturnType != "" {
			buf.WriteString(" " + opts.ReturnType)
		}

		buf.WriteString(" {\n")

		// Add body
		if opts.Body != "" {
			// Split body into lines and indent
			lines := strings.Split(opts.Body, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					buf.WriteString("\t" + line + "\n")
				} else {
					buf.WriteString("\n")
				}
			}
		}

		buf.WriteString("}\n")
	}

	// Format the code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), nil // Return unformatted if format fails
	}

	return formatted, nil
}

// createNewFileWithFunction creates a new Go file with the specified function
func createNewFileWithFunction(opts FunctionOpts, packageName string) ([]byte, error) {
	if packageName == "" {
		packageName = "main"
	}

	var buf bytes.Buffer

	// Write package declaration
	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// If Content is provided, use it directly
	if opts.Content != "" {
		buf.WriteString(fmt.Sprintf("func %s%s\n", opts.Name, opts.Content))
	} else {
		// Write function
		buf.WriteString(fmt.Sprintf("func %s(", opts.Name))

		// Add parameters
		for i, param := range opts.Parameters {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(fmt.Sprintf("%s %s", param.Name, param.Type))
		}

		buf.WriteString(")")

		// Add return type
		if opts.ReturnType != "" {
			buf.WriteString(" " + opts.ReturnType)
		}

		buf.WriteString(" {\n")

		// Add body
		if opts.Body != "" {
			// Split body into lines and indent
			lines := strings.Split(opts.Body, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					buf.WriteString("\t" + line + "\n")
				} else {
					buf.WriteString("\n")
				}
			}
		}

		buf.WriteString("}\n")
	}

	// Format the code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), nil // Return unformatted if format fails
	}

	return formatted, nil
}

// createNewFileWithVariable creates a new Go file with the specified variables
func createNewFileWithVariable(opts VariableOpts, packageName string) ([]byte, error) {
	if packageName == "" {
		packageName = "main"
	}

	var buf bytes.Buffer

	// Write package declaration
	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// If Content is provided, use it directly
	if opts.Content != "" {
		buf.WriteString(opts.Content)
	} else {
		// Write variables
		for _, variable := range opts.Variables {
			buf.WriteString("var ")
			buf.WriteString(variable.Name)
			if variable.Type != "" {
				buf.WriteString(" " + variable.Type)
			}
			if variable.Value != "" {
				buf.WriteString(" = " + variable.Value)
			}
			buf.WriteString("\n")
		}
	}

	// Format the code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), nil // Return unformatted if format fails
	}

	return formatted, nil
}

// createNewFileWithConstant creates a new Go file with the specified constants
func createNewFileWithConstant(opts ConstantOpts, packageName string) ([]byte, error) {
	if packageName == "" {
		packageName = "main"
	}

	var buf bytes.Buffer

	// Write package declaration
	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// If Content is provided, use it directly
	if opts.Content != "" {
		buf.WriteString(opts.Content)
	} else {
		// Write constants
		for _, constant := range opts.Constants {
			buf.WriteString("const ")
			buf.WriteString(constant.Name)
			if constant.Type != "" {
				buf.WriteString(" " + constant.Type)
			}
			if constant.Value != "" {
				buf.WriteString(" = " + constant.Value)
			}
			buf.WriteString("\n")
		}
	}

	// Format the code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), nil // Return unformatted if format fails
	}

	return formatted, nil
}

// createNewFileWithType creates a new Go file with the specified type definitions
func createNewFileWithType(opts TypeOpts, packageName string) ([]byte, error) {
	if packageName == "" {
		packageName = "main"
	}

	var buf bytes.Buffer

	// Write package declaration
	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// If Content is provided, use it directly
	if opts.Content != "" {
		buf.WriteString(opts.Content)
	} else {
		// Write types
		for _, typeDef := range opts.Types {
			buf.WriteString("type ")
			buf.WriteString(typeDef.Name)
			buf.WriteString(" ")
			buf.WriteString(typeDef.Definition)
			buf.WriteString("\n")
		}
	}

	// Format the code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), nil // Return unformatted if format fails
	}

	return formatted, nil
}

// modifyExistingFileForMethod modifies an existing Go file to ensure/modify methods
func modifyExistingFileForMethod(content []byte, opts MethodOpts) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	var methodCode []byte
	if opts.Content != "" {
		// Use Content directly with receiver and method name
		receiverName := opts.ReceiverName
		if receiverName == "" {
			// Default receiver name (first letter of type, skipping pointer)
			typeWithoutPointer := strings.TrimPrefix(opts.ReceiverType, "*")
			if len(typeWithoutPointer) > 0 {
				receiverName = strings.ToLower(string(typeWithoutPointer[0]))
			} else {
				receiverName = "r"
			}
		}
		methodCode = []byte(fmt.Sprintf("package tmp\n\nfunc (%s %s) %s%s", receiverName, opts.ReceiverType, opts.Name, opts.Content))
	} else {
		// For now, just append the method - this is a basic implementation
		// A full implementation would parse existing methods and merge/replace them
		var err error
		methodCode, err = createMethodDeclaration(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to create method: %w", err)
		}
	}

	// Parse the method code and add to the file
	methodFile, err := parser.ParseFile(fset, "", methodCode, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse method code: %w", err)
	}

	// Append the method declaration to the file
	for _, decl := range methodFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			file.Decls = append(file.Decls, funcDecl)
		}
	}

	// Format and return the modified code
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("failed to format Go code: %w", err)
	}

	return buf.Bytes(), nil
}

// modifyExistingFileForFunction modifies an existing Go file to ensure/modify functions
func modifyExistingFileForFunction(content []byte, opts FunctionOpts) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	var functionCode []byte
	if opts.Content != "" {
		// Use Content directly with function name
		functionCode = []byte(fmt.Sprintf("package tmp\n\nfunc %s%s", opts.Name, opts.Content))
	} else {
		// For now, just append the function - this is a basic implementation
		var err error
		functionCode, err = createFunctionDeclaration(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to create function: %w", err)
		}
	}

	// Parse the function code and add to the file
	functionFile, err := parser.ParseFile(fset, "", functionCode, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse function code: %w", err)
	}

	// Append the function declaration to the file
	for _, decl := range functionFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			file.Decls = append(file.Decls, funcDecl)
		}
	}

	// Format and return the modified code
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("failed to format Go code: %w", err)
	}

	return buf.Bytes(), nil
}

// modifyExistingFileForVariable modifies an existing Go file to ensure/modify variables
func modifyExistingFileForVariable(content []byte, opts VariableOpts) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	// If Content is provided, parse and append it
	if opts.Content != "" {
		// Parse the content as Go code
		contentWithPackage := fmt.Sprintf("package tmp\n\n%s", opts.Content)
		contentFile, err := parser.ParseFile(fset, "", contentWithPackage, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content: %w", err)
		}
		// Add the declarations from content
		for _, decl := range contentFile.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
				file.Decls = append(file.Decls, genDecl)
			}
		}
	} else {
		// For now, just append variables - this is a basic implementation
		for _, variable := range opts.Variables {
			varDecl := createVariableDeclaration(variable)
			file.Decls = append(file.Decls, varDecl)
		}
	}

	// Format and return the modified code
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("failed to format Go code: %w", err)
	}

	return buf.Bytes(), nil
}

// modifyExistingFileForConstant modifies an existing Go file to ensure/modify constants
func modifyExistingFileForConstant(content []byte, opts ConstantOpts) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	// If Content is provided, parse and append it
	if opts.Content != "" {
		// Parse the content as Go code
		contentWithPackage := fmt.Sprintf("package tmp\n\n%s", opts.Content)
		contentFile, err := parser.ParseFile(fset, "", contentWithPackage, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content: %w", err)
		}
		// Add the declarations from content
		for _, decl := range contentFile.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
				file.Decls = append(file.Decls, genDecl)
			}
		}
	} else {
		// For now, just append constants - this is a basic implementation
		for _, constant := range opts.Constants {
			constDecl := createConstantDeclaration(constant)
			file.Decls = append(file.Decls, constDecl)
		}
	}

	// Format and return the modified code
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("failed to format Go code: %w", err)
	}

	return buf.Bytes(), nil
}

// modifyExistingFileForType modifies an existing Go file to ensure/modify type definitions
func modifyExistingFileForType(content []byte, opts TypeOpts) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	// If Content is provided, parse and append it
	if opts.Content != "" {
		// Parse the content as Go code
		contentWithPackage := fmt.Sprintf("package tmp\n\n%s", opts.Content)
		contentFile, err := parser.ParseFile(fset, "", contentWithPackage, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content: %w", err)
		}
		// Add the declarations from content
		for _, decl := range contentFile.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
				file.Decls = append(file.Decls, genDecl)
			}
		}
	} else {
		// For now, just append types - this is a basic implementation
		for _, typeDef := range opts.Types {
			typeDecl := createTypeDeclaration(typeDef)
			file.Decls = append(file.Decls, typeDecl)
		}
	}

	// Format and return the modified code
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("failed to format Go code: %w", err)
	}

	return buf.Bytes(), nil
}

// Helper functions to create AST declarations

// createMethodDeclaration creates the Go code for a method
func createMethodDeclaration(opts MethodOpts) ([]byte, error) {
	var buf bytes.Buffer

	receiverName := opts.ReceiverName
	if receiverName == "" {
		// Default receiver name (first letter of type, skipping pointer)
		typeWithoutPointer := strings.TrimPrefix(opts.ReceiverType, "*")
		if len(typeWithoutPointer) > 0 {
			receiverName = strings.ToLower(string(typeWithoutPointer[0]))
		} else {
			receiverName = "r"
		}
	}

	buf.WriteString(fmt.Sprintf("package tmp\n\nfunc (%s %s) %s(", receiverName, opts.ReceiverType, opts.Name))

	// Add parameters
	for i, param := range opts.Parameters {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%s %s", param.Name, param.Type))
	}

	buf.WriteString(")")

	// Add return type
	if opts.ReturnType != "" {
		buf.WriteString(" " + opts.ReturnType)
	}

	buf.WriteString(" {\n")

	// Add body
	if opts.Body != "" {
		// Split body into lines and indent
		lines := strings.Split(opts.Body, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				buf.WriteString("\t" + line + "\n")
			} else {
				buf.WriteString("\n")
			}
		}
	}

	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

// createFunctionDeclaration creates the Go code for a function
func createFunctionDeclaration(opts FunctionOpts) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("package tmp\n\nfunc %s(", opts.Name))

	// Add parameters
	for i, param := range opts.Parameters {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%s %s", param.Name, param.Type))
	}

	buf.WriteString(")")

	// Add return type
	if opts.ReturnType != "" {
		buf.WriteString(" " + opts.ReturnType)
	}

	buf.WriteString(" {\n")

	// Add body
	if opts.Body != "" {
		// Split body into lines and indent
		lines := strings.Split(opts.Body, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				buf.WriteString("\t" + line + "\n")
			} else {
				buf.WriteString("\n")
			}
		}
	}

	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

// createVariableDeclaration creates an AST variable declaration
func createVariableDeclaration(variable Variable) *ast.GenDecl {
	var specs []ast.Spec

	spec := &ast.ValueSpec{
		Names: []*ast.Ident{{Name: variable.Name}},
	}

	if variable.Type != "" {
		spec.Type = parseTypeExpr(variable.Type)
	}

	if variable.Value != "" {
		// This is a simplified value parsing - in reality we would need to parse expressions properly
		spec.Values = []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: variable.Value}}
	}

	specs = append(specs, spec)

	return &ast.GenDecl{
		Tok:   token.VAR,
		Specs: specs,
	}
}

// createConstantDeclaration creates an AST constant declaration
func createConstantDeclaration(constant Constant) *ast.GenDecl {
	var specs []ast.Spec

	spec := &ast.ValueSpec{
		Names: []*ast.Ident{{Name: constant.Name}},
	}

	if constant.Type != "" {
		spec.Type = parseTypeExpr(constant.Type)
	}

	if constant.Value != "" {
		// This is a simplified value parsing - in reality we would need to parse expressions properly
		spec.Values = []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: constant.Value}}
	}

	specs = append(specs, spec)

	return &ast.GenDecl{
		Tok:   token.CONST,
		Specs: specs,
	}
}

// createTypeDeclaration creates an AST type declaration
func createTypeDeclaration(typeDef TypeDef) *ast.GenDecl {
	var specs []ast.Spec

	spec := &ast.TypeSpec{
		Name: &ast.Ident{Name: typeDef.Name},
		Type: parseTypeExpr(typeDef.Definition),
	}

	specs = append(specs, spec)

	return &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: specs,
	}
}
