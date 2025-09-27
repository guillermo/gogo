package gogo

import (
	"fmt"

	"github.com/guillermo/gogo/gogotest"
)

// ExampleProject_Struct_withFields demonstrates using the new Struct API with Fields
func ExampleProject_Struct_withFields() {
	fs := gogotest.New("")

	project, _ := New(Options{
		FS:                 fs,
		ConflictFunc:       ConflictAccept,
		InitialPackageName: "main",
	})

	// Create a User struct using the Fields approach
	project.Struct(StructOpts{
		Filename: "user.go",
		Name:     "User",
		Fields: []StructField{
			{Name: "ID", Type: "string", Annotation: `json:"id"`},
			{Name: "Name", Type: "string", Annotation: `json:"name"`},
			{Name: "Email", Type: "string", Annotation: `json:"email"`},
		},
		PreserveExisting: true,
	})

	fmt.Println(fs)

	// Output:
	// # user.go
	// package main
	//
	// type User struct {
	// 	ID    string `json:"id"`
	// 	Name  string `json:"name"`
	// 	Email string `json:"email"`
	// }
}

// ExampleProject_Struct_withContent demonstrates using the new Struct API with Content
func ExampleProject_Struct_withContent() {
	fs := gogotest.New("")

	project, _ := New(Options{
		FS:                 fs,
		ConflictFunc:       ConflictAccept,
		InitialPackageName: "models",
	})

	// Create a Product struct using the Content approach
	project.Struct(StructOpts{
		Filename: "product.go",
		Name:     "Product",
		Content: `
			ID    string  ` + "`json:\"id\"`" + `
			Title string  ` + "`json:\"title\"`" + `
			Price float64 ` + "`json:\"price\"`" + `
		`,
	})

	fmt.Println(fs)

	// Output:
	// # product.go
	// package models
	//
	// type Product struct {
	// 	ID    string  `json:"id"`
	// 	Title string  `json:"title"`
	// 	Price float64 `json:"price"`
	// }
}

// ExampleProject_Method demonstrates creating a method
func ExampleProject_Method() {
	fs := gogotest.New("")

	project, _ := New(Options{
		FS:                 fs,
		ConflictFunc:       ConflictAccept,
		InitialPackageName: "main",
	})

	// Create a method for a User struct
	project.Method(MethodOpts{
		Filename:     "user.go",
		Name:         "String",
		ReceiverName: "u",
		ReceiverType: "*User",
		ReturnType:   "string",
		Body:         `return fmt.Sprintf("User{ID: %s, Name: %s}", u.ID, u.Name)`,
	})

	fmt.Println(fs)

	// Output:
	// # user.go
	// package main
	//
	// func (u *User) String() string {
	// 	return fmt.Sprintf("User{ID: %s, Name: %s}", u.ID, u.Name)
	// }
}

// ExampleProject_Function demonstrates creating a function
func ExampleProject_Function() {
	fs := gogotest.New("")

	project, _ := New(Options{
		FS:                 fs,
		ConflictFunc:       ConflictAccept,
		InitialPackageName: "utils",
	})

	// Create a utility function
	project.Function(FunctionOpts{
		Filename:   "helpers.go",
		Name:       "StringToUpper",
		Parameters: []Parameter{{Name: "s", Type: "string"}},
		ReturnType: "string",
		Body:       `return strings.ToUpper(s)`,
	})

	fmt.Println(fs)

	// Output:
	// # helpers.go
	// package utils
	//
	// func StringToUpper(s string) string {
	// 	return strings.ToUpper(s)
	// }
}

// ExampleProject_Variable demonstrates creating variables
func ExampleProject_Variable() {
	fs := gogotest.New("")

	project, _ := New(Options{
		FS:                 fs,
		ConflictFunc:       ConflictAccept,
		InitialPackageName: "config",
	})

	// Create configuration variables
	project.Variable(VariableOpts{
		Filename: "config.go",
		Variables: []Variable{
			{Name: "DefaultTimeout", Type: "time.Duration", Value: "30 * time.Second"},
			{Name: "MaxRetries", Type: "int", Value: "3"},
		},
	})

	fmt.Println(fs)

	// Output:
	// # config.go
	// package config
	//
	// var DefaultTimeout time.Duration = 30 * time.Second
	// var MaxRetries int = 3
}

// ExampleProject_Constant demonstrates creating constants
func ExampleProject_Constant() {
	fs := gogotest.New("")

	project, _ := New(Options{
		FS:                 fs,
		ConflictFunc:       ConflictAccept,
		InitialPackageName: "main",
	})

	// Create constants
	project.Constant(ConstantOpts{
		Filename: "constants.go",
		Constants: []Constant{
			{Name: "Version", Type: "string", Value: `"1.0.0"`},
			{Name: "MaxUsers", Type: "int", Value: "1000"},
		},
	})

	fmt.Println(fs)

	// Output:
	// # constants.go
	// package main
	//
	// const Version string = "1.0.0"
	// const MaxUsers int = 1000
}

// ExampleProject_Type demonstrates creating type definitions
func ExampleProject_Type() {
	fs := gogotest.New("")

	project, _ := New(Options{
		FS:                 fs,
		ConflictFunc:       ConflictAccept,
		InitialPackageName: "types",
	})

	// Create type definitions
	project.Type(TypeOpts{
		Filename: "types.go",
		Types: []TypeDef{
			{Name: "UserID", Definition: "string"},
			{Name: "Handler", Definition: "func(ctx context.Context) error"},
		},
	})

	fmt.Println(fs)

	// Output:
	// # types.go
	// package types
	//
	// type UserID string
	// type Handler func(ctx context.Context) error
}
