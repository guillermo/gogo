package gogo

import "go/ast"

// copyDecl creates a deep copy of a declaration.
// It handles GenDecl and FuncDecl types with their nested elements.
func copyDecl(decl ast.Decl) ast.Decl {
	switch d := decl.(type) {
	case *ast.GenDecl:
		genDecl := &ast.GenDecl{
			Tok:    d.Tok,
			Lparen: d.Lparen,
			Rparen: d.Rparen,
			Doc:    d.Doc,
		}
		specs := make([]ast.Spec, len(d.Specs))
		for i, spec := range d.Specs {
			specs[i] = copySpec(spec)
		}
		genDecl.Specs = specs
		return genDecl
	case *ast.FuncDecl:
		funcDecl := &ast.FuncDecl{
			Name: ast.NewIdent(d.Name.Name),
			Doc:  d.Doc,
			Type: copyFuncType(d.Type),
			Body: copyBlockStmt(d.Body),
		}
		if d.Recv != nil {
			funcDecl.Recv = copyFieldList(d.Recv)
		}
		return funcDecl
	default:
		return d
	}
}

// copySpec creates a deep copy of a spec.
// It handles TypeSpec, ValueSpec, and ImportSpec types.
func copySpec(spec ast.Spec) ast.Spec {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		return &ast.TypeSpec{
			Name:    ast.NewIdent(s.Name.Name),
			Type:    copyExpr(s.Type),
			Doc:     s.Doc,
			Comment: s.Comment,
		}
	case *ast.ValueSpec:
		names := make([]*ast.Ident, len(s.Names))
		for i, name := range s.Names {
			names[i] = ast.NewIdent(name.Name)
		}
		values := make([]ast.Expr, len(s.Values))
		for i, value := range s.Values {
			values[i] = copyExpr(value)
		}
		return &ast.ValueSpec{
			Names:   names,
			Type:    copyExpr(s.Type),
			Values:  values,
			Doc:     s.Doc,
			Comment: s.Comment,
		}
	case *ast.ImportSpec:
		return &ast.ImportSpec{
			Name:    s.Name,
			Path:    s.Path,
			Doc:     s.Doc,
			Comment: s.Comment,
		}
	default:
		return s
	}
}

// copyExpr creates a deep copy of an expression.
// It handles various expression types like StructType, Ident, StarExpr, etc.
// If the expression is nil, it returns nil.
func copyExpr(expr ast.Expr) ast.Expr {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.StructType:
		return &ast.StructType{
			Fields:     copyFieldList(e.Fields),
			Incomplete: e.Incomplete,
		}
	case *ast.Ident:
		return ast.NewIdent(e.Name)
	case *ast.StarExpr:
		return &ast.StarExpr{
			X: copyExpr(e.X),
		}
	case *ast.ArrayType:
		return &ast.ArrayType{
			Len: copyExpr(e.Len),
			Elt: copyExpr(e.Elt),
		}
	case *ast.SelectorExpr:
		return &ast.SelectorExpr{
			X:   copyExpr(e.X),
			Sel: ast.NewIdent(e.Sel.Name),
		}
	case *ast.MapType:
		return &ast.MapType{
			Key:   copyExpr(e.Key),
			Value: copyExpr(e.Value),
		}
	case *ast.InterfaceType:
		return &ast.InterfaceType{
			Methods:    copyFieldList(e.Methods),
			Incomplete: e.Incomplete,
		}
	case *ast.FuncType:
		return copyFuncType(e)
	default:
		return e
	}
}

// copyFieldList creates a deep copy of a field list.
// It returns nil if the input list is nil.
func copyFieldList(list *ast.FieldList) *ast.FieldList {
	if list == nil {
		return nil
	}

	fieldList := &ast.FieldList{}
	if list.List != nil {
		fields := make([]*ast.Field, len(list.List))
		for i, field := range list.List {
			fields[i] = copyField(field)
		}
		fieldList.List = fields
	}
	return fieldList
}

// copyField creates a deep copy of a field.
// It copies Names, Type, Tag, Doc, and Comment fields.
func copyField(field *ast.Field) *ast.Field {
	names := make([]*ast.Ident, len(field.Names))
	for i, name := range field.Names {
		names[i] = ast.NewIdent(name.Name)
	}

	newField := &ast.Field{
		Names:   names,
		Type:    copyExpr(field.Type),
		Tag:     field.Tag,
		Doc:     field.Doc,
		Comment: field.Comment,
	}
	return newField
}

// copyBlockStmt creates a deep copy of a block statement.
// It returns nil if the input block is nil.
func copyBlockStmt(block *ast.BlockStmt) *ast.BlockStmt {
	if block == nil {
		return nil
	}

	stmts := make([]ast.Stmt, len(block.List))
	for i, stmt := range block.List {
		stmts[i] = copyStmt(stmt)
	}

	return &ast.BlockStmt{
		Lbrace: block.Lbrace,
		List:   stmts,
		Rbrace: block.Rbrace,
	}
}

// copyFuncType creates a deep copy of a function type.
// It preserves Params and Results fields.
func copyFuncType(funcType *ast.FuncType) *ast.FuncType {
	return &ast.FuncType{
		Params:  copyFieldList(funcType.Params),
		Results: copyFieldList(funcType.Results),
	}
}

// copyStmt creates a deep copy of a statement.
// It handles various statement types and returns nil if the input statement is nil.
func copyStmt(stmt ast.Stmt) ast.Stmt {
	if stmt == nil {
		return nil
	}

	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		results := make([]ast.Expr, len(s.Results))
		for i, result := range s.Results {
			results[i] = copyExpr(result)
		}
		return &ast.ReturnStmt{
			Return:  s.Return,
			Results: results,
		}
	case *ast.AssignStmt:
		lhs := make([]ast.Expr, len(s.Lhs))
		for i, expr := range s.Lhs {
			lhs[i] = copyExpr(expr)
		}
		rhs := make([]ast.Expr, len(s.Rhs))
		for i, expr := range s.Rhs {
			rhs[i] = copyExpr(expr)
		}
		return &ast.AssignStmt{
			Lhs:    lhs,
			Rhs:    rhs,
			Tok:    s.Tok,
			TokPos: s.TokPos,
		}
	case *ast.ExprStmt:
		return &ast.ExprStmt{
			X: copyExpr(s.X),
		}
	default:
		return s
	}
}
