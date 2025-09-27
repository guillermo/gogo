package tests

import (
	"testing"

	"github.com/guillermo/gogo"
	"github.com/guillermo/gogo/gogotest"
)

func TestSimple(t *testing.T) {
	// Test that gogotest works with gogo
	fs := gogotest.New("")

	project, err := gogo.New(gogo.Options{
		FS:                 fs,
		ConflictFunc:       gogo.ConflictAccept,
		InitialPackageName: "testpkg",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = project.Struct(gogo.StructOpts{
		Filename: "user.go",
		Name:     "User",
		Fields: []gogo.StructField{
			{Name: "ID", Type: "string", Annotation: `json:"id"`},
			{Name: "Name", Type: "string", Annotation: `json:"name"`},
		},
		PreserveExisting: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test that we can assert on the results
	err = fs.Assert(`type User struct`)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Assert(`ID string`)
	if err != nil {
		t.Fatal(err)
	}

	// Test Method functionality
	err = project.Method(gogo.MethodOpts{
		Filename:     "user.go",
		Name:         "GetID",
		ReceiverName: "u",
		ReceiverType: "*User",
		ReturnType:   "string",
		Body:         `return u.ID`,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test Function functionality
	err = project.Function(gogo.FunctionOpts{
		Filename:   "helpers.go",
		Name:       "NewUser",
		Parameters: []gogo.Parameter{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}},
		ReturnType: "*User",
		Body:       `return &User{ID: id, Name: name}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test Variable functionality
	err = project.Variable(gogo.VariableOpts{
		Filename: "config.go",
		Variables: []gogo.Variable{
			{Name: "DefaultName", Type: "string", Value: `"Anonymous"`},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test Constant functionality
	err = project.Constant(gogo.ConstantOpts{
		Filename: "constants.go",
		Constants: []gogo.Constant{
			{Name: "MaxAge", Type: "int", Value: "120"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test Type functionality
	err = project.Type(gogo.TypeOpts{
		Filename: "types.go",
		Types: []gogo.TypeDef{
			{Name: "UserID", Definition: "string"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test assertions for the new constructs
	err = fs.Assert(`func (u *User) GetID() string`)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Assert(`func NewUser(id string, name string) *User`)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Assert(`var DefaultName string = "Anonymous"`)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Assert(`const MaxAge int = 120`)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Assert(`type UserID string`)
	if err != nil {
		t.Fatal(err)
	}
}
