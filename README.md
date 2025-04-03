# GoGo - Go to Go code manipulator

[![Go Report Card](https://goreportcard.com/badge/github.com/guillermo/gogo)](https://goreportcard.com/report/github.com/guillermo/gogo)
[![GoDoc](https://godoc.org/github.com/guillermo/gogo?status.svg)](https://godoc.org/github.com/guillermo/gogo)
[![Go Version](https://img.shields.io/github/go-mod/go-version/guillermo/gogo)](https://github.com/guillermo/gogo)
[![License](https://img.shields.io/github/license/guillermo/gogo)](https://github.com/guillermo/gogo/blob/main/LICENSE)
[![Build Status](https://github.com/guillermo/gogo/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/guillermo/gogo/actions/workflows/go.yml)

![GoGo Logo](./gogo.png)


## Features

GoGo allows you to:

* Parse a directory containing Go files and a go.mod file
* Manipulate the parsed code:
  * Rename the package
  * Remove, duplicate, or rename files
  * Rename types throughout the codebase
  * Modify struct fields (add, remove, change types and tags)
* Write the modified code to a new directory

## Installation

```bash
go get github.com/guillermo/gogo
```

## Quick Start

```go
template := gogo.Open("./example")
template.PackageName = "mypackage"
file, _ := template.ExtractAndRemove("user.go")
for _, name := range []string{"Car","Toy", "Person"} {
  file.RenameType("User", name)  // This will rename all types called User to the name
  file.OpenStruct(name, func(s gogo.Struct){
    s.Add("Name", "string", map[string]string{"json":"name", "db": "name_col"})
    s.Remove("Id")
  })
  template.Add(name + ".go", file) // this adds the new file to the template
}
template.Write("./generated")
```

## API Reference

### Open

```go
func Open(dir string) (*Template, error)
```

Open reads all the Go files in the given directory and returns a Template object.

### Template

```go
type Template struct {
  PackageName string
  Files       map[string]*File
}
```

Template represents a Go package template that can be manipulated.

#### Methods

- `ExtractAndRemove(name string) (*File, error)`: Extracts a file from the template and removes it
- `Add(name string, file *File)`: Adds a file to the template
- `Write(dir string) error`: Writes the template to a directory
- `OpenStruct(name string, modifier StructModifier) error`: Finds a struct by name across all files and applies the modifier function

### File

```go
type File struct {
  Name      string
  AstFile   *ast.File
  Template  *Template
}
```

File represents a Go file that can be manipulated.

#### Methods

- `RenameType(oldName, newName string) error`: Renames types throughout the file
- `OpenStruct(name string, modifier StructModifier) error`: Finds a struct by name and applies the modifier function
- `Clone() *File`: Creates a deep copy of the file
- `RenameFunction(oldName, newName string) error`: Renames a function in the file
- `RemoveFunction(name string) error`: Removes a function from the file
- `DuplicateFunction(oldName, newName string) error`: Duplicates a function with a new name

### Field

```go
type Field struct {
  Name     string
  Type     string
  Tags     map[string]string
  AstField *ast.Field
}
```

Field represents a struct field with a more convenient interface than the ast.Field type.

#### Methods

- `SetType(typeName string)`: Changes the type of the field
- `SetTags(tags map[string]string)`: Sets all tags for the field
- `AddTag(key, value string)`: Adds or updates a single tag
- `String() string`: Returns a string representation of the field

### Struct

```go
type Struct struct {
  Name        string
  Fields      []*ast.Field
  StructType  *ast.StructType
  Parent      *File
  Declaration *ast.TypeSpec
}
```

Struct represents a Go struct that can be manipulated.

#### Methods

- `Add(name, fieldType string, tags map[string]string)`: Adds a field to the struct
- `Remove(name string) error`: Removes a field from the struct
- `Field(name string) *Field`: Finds a field by name and returns a Field representation
- `AddField(field *Field)`: Adds a field to the struct using a Field object
- `RemoveMethod(methodName string) error`: Removes a method from the struct
- `DuplicateMethod(oldName, newName string) error`: Duplicates a method with a new name
- `RenameMethod(oldName, newName string) error`: Renames a method of the struct

### Utility Functions

- `NewField(name, fieldType string, tags map[string]string) *Field`: Creates a new Field object

## Examples

See the [example](example) directory for a sample Go file and the [cmd/example](cmd/example) directory for a complete example of how to use GoGo.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.