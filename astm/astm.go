/* Package astm allows to modify a go abstract syntax tree

// Parse parse a set of bytes
code := astm.Parse(ast)

// Package set the name of the packages
code.Package("newName")

// Rename renames a function, struct, variable, or constant.
// If there are any references to the identifier, those references are also renamed.
// For example, renaming a struct will also rename all its usages in variable declarations,
// function parameters, and return types.
err := code.Rename("Before", "After")

// Set adds new code to the tree. The code can be:
// - A function declaration
// - A variable declaration
// - A constant declaration
// - A struct declaration
// The code must be valid Go code and include the package declaration.
code.Set([]byte(`package main
func (a After) Sum() int { return 42 }
const PI = 3.14
var globalVar = "value"
type NewStruct struct { Field int }`))

// Delete removes a function, struct, variable, or constant.
// It will fail if there are any references to the identifier.
// For example, deleting a struct that is used in variable declarations
// or function parameters will fail.
err := code.Delete("FuncVarConst")

// FieldSet sets a field in a struct
code.FieldSet("Struct", "FiledName", "int", "tags")

// FieldRename renames a field. If there is any reference, it will also rename it.
code.FieldRename("Struct", "Before", "After")

// FieldDelete deletes a field
code.FieldDelete("Struct", "asdf")

// WriteTo writes the code to the destination
code.WriteTo(w)
*/

package astm

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"strings"
)

// Code represents a Go source code file that can be modified
type Code struct {
	fset *token.FileSet
	file *ast.File
}

// Parse parses the Go source code and returns a Code object
func Parse(src []byte) *Code {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil
	}
	return &Code{fset: fset, file: file}
}

// WriteTo writes the modified code to the given writer
func (c *Code) WriteTo(w io.Writer) error {
	return printer.Fprint(w, c.fset, c.file)
}

// Package sets the package name
func (c *Code) Package(name string) {
	c.file.Name.Name = name
}

// Rename renames an identifier and its references
func (c *Code) Rename(old, new string) error {
	// First, find all declarations of the identifier
	var decls []ast.Node
	ast.Inspect(c.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if x.Name.Name == old {
				decls = append(decls, x)
			}
		case *ast.FuncDecl:
			if x.Name.Name == old {
				decls = append(decls, x)
			}
		case *ast.ValueSpec:
			for _, name := range x.Names {
				if name.Name == old {
					decls = append(decls, x)
					break
				}
			}
		}
		return true
	})

	if len(decls) == 0 {
		return nil // No declarations found, nothing to rename
	}

	// Then, find and rename all references
	ast.Inspect(c.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if x.Name == old {
				x.Name = new
			}
		}
		return true
	})

	// Finally, rename the declarations
	for _, decl := range decls {
		switch x := decl.(type) {
		case *ast.TypeSpec:
			x.Name.Name = new
		case *ast.FuncDecl:
			x.Name.Name = new
		case *ast.ValueSpec:
			for _, name := range x.Names {
				if name.Name == old {
					name.Name = new
				}
			}
		}
	}

	return nil
}

// Set adds new code to the tree
func (c *Code) Set(src []byte) error {
	// Parse the new code
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return err
	}

	// Add all declarations from the new code to the existing file
	for _, decl := range file.Decls {
		c.file.Decls = append(c.file.Decls, decl)
	}

	return nil
}

// Delete removes an identifier and its references
func (c *Code) Delete(name string) error {
	// First, find all references to the identifier
	var refs []*ast.Ident
	ast.Inspect(c.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if x.Name == name {
				refs = append(refs, x)
			}
		}
		return true
	})

	// Then, find the declaration
	var decl interface{}
	var declIdent *ast.Ident
	for _, ref := range refs {
		if ref.Obj != nil && ref.Obj.Decl != nil {
			switch d := ref.Obj.Decl.(type) {
			case *ast.FuncDecl:
				if d.Name == ref {
					decl = d
					declIdent = ref
					break
				}
			case *ast.TypeSpec:
				if d.Name == ref {
					decl = d
					declIdent = ref
					break
				}
			case *ast.ValueSpec:
				for _, n := range d.Names {
					if n == ref {
						decl = d
						declIdent = ref
						break
					}
				}
			}
		}
	}

	if decl == nil {
		return nil // Nothing to delete
	}

	// Check if there are any references other than the declaration
	var hasRefs bool
	for _, ref := range refs {
		if ref == declIdent {
			continue // Skip the declaration
		}
		hasRefs = true
		break
	}

	if hasRefs {
		return fmt.Errorf("cannot delete %s: it has references", name)
	}

	// Finally, remove the declaration
	var newDecls []ast.Decl
	for _, d := range c.file.Decls {
		keep := true
		switch x := d.(type) {
		case *ast.FuncDecl:
			if x == decl {
				keep = false
			}
		case *ast.GenDecl:
			var newSpecs []ast.Spec
			for _, spec := range x.Specs {
				keepSpec := true
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s == decl {
						keepSpec = false
					}
				case *ast.ValueSpec:
					if s == decl {
						keepSpec = false
					}
				}
				if keepSpec {
					newSpecs = append(newSpecs, spec)
				}
			}
			if len(newSpecs) == 0 {
				keep = false
			} else {
				x.Specs = newSpecs
			}
		}
		if keep {
			newDecls = append(newDecls, d)
		}
	}
	c.file.Decls = newDecls

	return nil
}

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

// BuildTags sets the build tags for the file
func (c *Code) BuildTags(tags []string) {
	if len(tags) == 0 {
		return
	}

	// Create a new comment group for build tags
	comment := "//go:build " + strings.Join(tags, " && ")
	c.file.Comments = append([]*ast.CommentGroup{
		{
			List: []*ast.Comment{
				{
					Text: comment,
				},
			},
		},
	}, c.file.Comments...)
}
