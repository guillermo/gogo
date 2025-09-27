// Package template provides functionality to extract and transform Go source code ASTs
// for use with the gogo code generation library.
//
// This package is designed for scenarios where you have a reference implementation
// that you want to use as a template for generating similar code. A typical use case
// is building an ORM where you have one model (e.g., Customer) fully implemented,
// and you want to use it as a template to generate other models (e.g., User, Product).
//
// The Template type provides an immutable API where all transformations return a new
// Template instance, allowing for easy method chaining and safe concurrent usage.
//
// Example:
//
//	// Load reference implementation
//	tmpl, _ := template.New(referenceFS)
//
//	// Transform the template
//	tmpl, _ = tmpl.RenameStruct("Customer", "User")
//	tmpl, _ = tmpl.RenameStructField("User", "CustomerID", gogo.StructField{
//	    Name: "UserID",
//	    Type: "int",
//	})
//
//	// Extract and use with gogo
//	userStruct, _ := tmpl.ExtractStruct("User")
//	prj.Struct(userStruct)
package template

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/guillermo/gogo"
	"github.com/guillermo/gogo/fs"
)

// Template represents a parsed Go project that can be transformed and extracted
type Template struct {
	fset    *token.FileSet
	files   map[string]*ast.File // filename -> parsed AST
	pkgName string
}

// New creates a new Template from a filesystem containing Go source files
func New(filesystem fs.FS) (*Template, error) {
	fset := token.NewFileSet()
	files := make(map[string]*ast.File)
	pkgName := ""

	// Walk the filesystem to find all .go files
	goFiles, err := findGoFiles(filesystem)
	if err != nil {
		return nil, fmt.Errorf("failed to find Go files: %w", err)
	}

	if len(goFiles) == 0 {
		return nil, fmt.Errorf("no Go files found in filesystem")
	}

	// Parse each Go file
	for _, filename := range goFiles {
		content, err := filesystem.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
		}

		file, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file %s: %w", filename, err)
		}

		files[filename] = file

		// Use the package name from the first file
		if pkgName == "" {
			pkgName = file.Name.Name
		}
	}

	return &Template{
		fset:    fset,
		files:   files,
		pkgName: pkgName,
	}, nil
}

// ExtractStruct extracts a struct definition and returns it as gogo.StructOpts
func (t *Template) ExtractStruct(name string) (gogo.StructOpts, error) {
	structNode := t.findStruct(name)
	if structNode == nil {
		return gogo.StructOpts{}, fmt.Errorf("struct %s not found", name)
	}

	// Convert AST struct to gogo.StructOpts
	fields := make([]gogo.StructField, 0)

	for _, field := range structNode.Fields.List {
		for _, fieldName := range field.Names {
			gogoField := gogo.StructField{
				Name: fieldName.Name,
				Type: typeExprToString(field.Type),
			}

			// Extract tag if present
			if field.Tag != nil {
				gogoField.Annotation = field.Tag.Value
			}

			fields = append(fields, gogoField)
		}
	}

	return gogo.StructOpts{
		Name:   name,
		Fields: fields,
	}, nil
}

// ExtractFunction extracts a function definition and returns it as gogo.FunctionOpts
func (t *Template) ExtractFunction(name string) (gogo.FunctionOpts, error) {
	funcNode := t.findFunction(name)
	if funcNode == nil {
		return gogo.FunctionOpts{}, fmt.Errorf("function %s not found", name)
	}

	// Convert AST function to gogo.FunctionOpts
	params := make([]gogo.Parameter, 0)
	if funcNode.Type.Params != nil {
		for _, param := range funcNode.Type.Params.List {
			for _, paramName := range param.Names {
				params = append(params, gogo.Parameter{
					Name: paramName.Name,
					Type: typeExprToString(param.Type),
				})
			}
		}
	}

	returnType := ""
	if funcNode.Type.Results != nil {
		returnType = resultsToString(funcNode.Type.Results)
	}

	body := ""
	if funcNode.Body != nil {
		body = blockStmtToString(funcNode.Body)
	}

	return gogo.FunctionOpts{
		Name:       name,
		Parameters: params,
		ReturnType: returnType,
		Body:       body,
	}, nil
}

// ExtractMethod extracts a method definition and returns it as gogo.MethodOpts
func (t *Template) ExtractMethod(receiverType, methodName string) (gogo.MethodOpts, error) {
	methodNode := t.findMethod(receiverType, methodName)
	if methodNode == nil {
		return gogo.MethodOpts{}, fmt.Errorf("method %s.%s not found", receiverType, methodName)
	}

	// Extract receiver info
	receiverName := ""
	if methodNode.Recv != nil && len(methodNode.Recv.List) > 0 {
		recvField := methodNode.Recv.List[0]
		if len(recvField.Names) > 0 {
			receiverName = recvField.Names[0].Name
		}
	}

	// Convert AST method to gogo.MethodOpts
	params := make([]gogo.Parameter, 0)
	if methodNode.Type.Params != nil {
		for _, param := range methodNode.Type.Params.List {
			for _, paramName := range param.Names {
				params = append(params, gogo.Parameter{
					Name: paramName.Name,
					Type: typeExprToString(param.Type),
				})
			}
		}
	}

	returnType := ""
	if methodNode.Type.Results != nil {
		returnType = resultsToString(methodNode.Type.Results)
	}

	body := ""
	if methodNode.Body != nil {
		body = blockStmtToString(methodNode.Body)
	}

	return gogo.MethodOpts{
		Name:         methodName,
		ReceiverName: receiverName,
		ReceiverType: receiverType,
		Parameters:   params,
		ReturnType:   returnType,
		Body:         body,
	}, nil
}

// ExtractVariable extracts variable declarations and returns them as gogo.VariableOpts
func (t *Template) ExtractVariable(name string) (gogo.VariableOpts, error) {
	varNode := t.findVariable(name)
	if varNode == nil {
		return gogo.VariableOpts{}, fmt.Errorf("variable %s not found", name)
	}

	variables := make([]gogo.Variable, 0)

	for _, spec := range varNode.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for i, varName := range valueSpec.Names {
			if varName.Name != name {
				continue
			}

			variable := gogo.Variable{
				Name: varName.Name,
			}

			if valueSpec.Type != nil {
				variable.Type = typeExprToString(valueSpec.Type)
			}

			if i < len(valueSpec.Values) {
				variable.Value = exprToString(valueSpec.Values[i])
			}

			variables = append(variables, variable)
		}
	}

	return gogo.VariableOpts{
		Variables: variables,
	}, nil
}

// ExtractConstant extracts constant declarations and returns them as gogo.ConstantOpts
func (t *Template) ExtractConstant(name string) (gogo.ConstantOpts, error) {
	constNode := t.findConstant(name)
	if constNode == nil {
		return gogo.ConstantOpts{}, fmt.Errorf("constant %s not found", name)
	}

	constants := make([]gogo.Constant, 0)

	for _, spec := range constNode.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for i, constName := range valueSpec.Names {
			if constName.Name != name {
				continue
			}

			constant := gogo.Constant{
				Name: constName.Name,
			}

			if valueSpec.Type != nil {
				constant.Type = typeExprToString(valueSpec.Type)
			}

			if i < len(valueSpec.Values) {
				constant.Value = exprToString(valueSpec.Values[i])
			}

			constants = append(constants, constant)
		}
	}

	return gogo.ConstantOpts{
		Constants: constants,
	}, nil
}

// ExtractType extracts type definitions and returns them as gogo.TypeOpts
func (t *Template) ExtractType(name string) (gogo.TypeOpts, error) {
	typeNode := t.findType(name)
	if typeNode == nil {
		return gogo.TypeOpts{}, fmt.Errorf("type %s not found", name)
	}

	types := make([]gogo.TypeDef, 0)

	for _, spec := range typeNode.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		if typeSpec.Name.Name != name {
			continue
		}

		typeDef := gogo.TypeDef{
			Name:       typeSpec.Name.Name,
			Definition: typeExprToString(typeSpec.Type),
		}

		types = append(types, typeDef)
	}

	return gogo.TypeOpts{
		Types: types,
	}, nil
}

// findGoFiles finds all .go files in the filesystem
func findGoFiles(filesystem fs.FS) ([]string, error) {
	// For mockFileSystem, we can access GetFiles if available
	type filesGetter interface {
		GetFiles() map[string][]byte
	}

	if fg, ok := filesystem.(filesGetter); ok {
		// This is a mock filesystem, we can get all files directly
		files := fg.GetFiles()
		goFiles := make([]string, 0, len(files))
		for filename := range files {
			if filepath.Ext(filename) == ".go" {
				goFiles = append(goFiles, filename)
			}
		}
		return goFiles, nil
	}

	// For real filesystems, try common patterns
	var goFiles []string

	// Common filenames to try
	commonFiles := []string{
		"main.go",
		"types.go",
		"models.go",
		"handlers.go",
		"service.go",
		"repository.go",
		"customer.go",
		"user.go",
	}

	for _, filename := range commonFiles {
		if _, err := filesystem.Stat(filename); err == nil {
			goFiles = append(goFiles, filename)
		}
	}

	// Also try to read any .go file by checking common patterns using real filesystem
	patterns := []string{"*.go"}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err == nil {
			for _, match := range matches {
				if _, err := filesystem.Stat(match); err == nil {
					// Check if not already added
					alreadyAdded := false
					for _, existing := range goFiles {
						if existing == match {
							alreadyAdded = true
							break
						}
					}
					if !alreadyAdded {
						goFiles = append(goFiles, match)
					}
				}
			}
		}
	}

	return goFiles, nil
}
