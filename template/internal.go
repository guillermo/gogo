package template

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
)

// findStruct finds a struct declaration by name across all files
func (t *Template) findStruct(name string) *ast.StructType {
	for _, file := range t.files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != name {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if ok {
					return structType
				}
			}
		}
	}
	return nil
}

// findStructGenDecl finds the GenDecl containing a struct by name
func (t *Template) findStructGenDecl(name string) (*ast.GenDecl, *ast.TypeSpec) {
	for _, file := range t.files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != name {
					continue
				}

				if _, ok := typeSpec.Type.(*ast.StructType); ok {
					return genDecl, typeSpec
				}
			}
		}
	}
	return nil, nil
}

// findFunction finds a function declaration by name
func (t *Template) findFunction(name string) *ast.FuncDecl {
	for _, file := range t.files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Recv != nil {
				continue // Skip methods
			}

			if funcDecl.Name.Name == name {
				return funcDecl
			}
		}
	}
	return nil
}

// findMethod finds a method declaration by receiver type and method name
func (t *Template) findMethod(receiverType, methodName string) *ast.FuncDecl {
	// Normalize receiver type (remove pointer if present)
	normalizedRecvType := strings.TrimPrefix(receiverType, "*")

	for _, file := range t.files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Recv == nil {
				continue // Skip functions
			}

			if funcDecl.Name.Name != methodName {
				continue
			}

			// Check receiver type
			if len(funcDecl.Recv.List) > 0 {
				recvType := typeExprToString(funcDecl.Recv.List[0].Type)
				normalizedRecv := strings.TrimPrefix(recvType, "*")

				if normalizedRecv == normalizedRecvType {
					return funcDecl
				}
			}
		}
	}
	return nil
}

// findVariable finds a variable declaration by name
func (t *Template) findVariable(name string) *ast.GenDecl {
	for _, file := range t.files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				continue
			}

			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for _, varName := range valueSpec.Names {
					if varName.Name == name {
						return genDecl
					}
				}
			}
		}
	}
	return nil
}

// findConstant finds a constant declaration by name
func (t *Template) findConstant(name string) *ast.GenDecl {
	for _, file := range t.files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.CONST {
				continue
			}

			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for _, constName := range valueSpec.Names {
					if constName.Name == name {
						return genDecl
					}
				}
			}
		}
	}
	return nil
}

// findType finds a type declaration by name
func (t *Template) findType(name string) *ast.GenDecl {
	for _, file := range t.files {
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
					return genDecl
				}
			}
		}
	}
	return nil
}

// typeExprToString converts an AST type expression to a string
func typeExprToString(expr ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), expr); err != nil {
		return ""
	}
	return buf.String()
}

// exprToString converts an AST expression to a string
func exprToString(expr ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), expr); err != nil {
		return ""
	}
	return buf.String()
}

// resultsToString converts function results to a string representation
func resultsToString(results *ast.FieldList) string {
	if results == nil || len(results.List) == 0 {
		return ""
	}

	var parts []string
	for _, field := range results.List {
		typeStr := typeExprToString(field.Type)
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				parts = append(parts, name.Name+" "+typeStr)
			}
		} else {
			parts = append(parts, typeStr)
		}
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return "(" + strings.Join(parts, ", ") + ")"
}

// blockStmtToString converts a block statement to a string (body of function/method)
func blockStmtToString(block *ast.BlockStmt) string {
	var buf bytes.Buffer

	// Format each statement in the block
	for _, stmt := range block.List {
		if err := format.Node(&buf, token.NewFileSet(), stmt); err != nil {
			continue
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// cloneTemplate creates a deep copy of the template
func (t *Template) cloneTemplate() *Template {
	newFiles := make(map[string]*ast.File)
	newFset := token.NewFileSet()

	for filename, file := range t.files {
		// Create a deep copy of the AST by formatting and re-parsing
		var buf bytes.Buffer
		if err := format.Node(&buf, t.fset, file); err != nil {
			continue
		}

		// Parse the formatted code to get a new AST
		newFile, err := parser.ParseFile(newFset, filename, buf.Bytes(), parser.ParseComments)
		if err != nil {
			continue
		}

		newFiles[filename] = newFile
	}

	return &Template{
		fset:    newFset,
		files:   newFiles,
		pkgName: t.pkgName,
	}
}

// renameIdentifier renames an identifier throughout all files
func (t *Template) renameIdentifier(oldName, newName string) {
	for _, file := range t.files {
		ast.Inspect(file, func(n ast.Node) bool {
			if ident, ok := n.(*ast.Ident); ok {
				if ident.Name == oldName {
					ident.Name = newName
				}
			}
			return true
		})
	}
}

// renameStructFieldIdentifier renames a struct field throughout all files
func (t *Template) renameStructFieldIdentifier(structName, oldFieldName, newFieldName string) {
	for _, file := range t.files {
		ast.Inspect(file, func(n ast.Node) bool {
			// Rename in struct definition
			if typeSpec, ok := n.(*ast.TypeSpec); ok && typeSpec.Name.Name == structName {
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					for _, field := range structType.Fields.List {
						for _, fieldName := range field.Names {
							if fieldName.Name == oldFieldName {
								fieldName.Name = newFieldName
							}
						}
					}
				}
			}

			// Rename in selector expressions (e.g., customer.CustomerID -> customer.UserID)
			if sel, ok := n.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == oldFieldName {
					// Check if the X part is of the struct type
					// This is a simplified check - a full implementation would need type information
					sel.Sel.Name = newFieldName
				}
			}

			return true
		})
	}
}
