package tests

import (
	"testing"

	"github.com/guillermo/gogo"
	"github.com/guillermo/gogo/gogotest"
)

func TestProjectStruct(t *testing.T) {
	t.Run("CreateNewStruct", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "models",
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
				{Name: "Email", Type: "string", Annotation: `json:"email"`},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify the struct was created
		if err := fs.Assert(`type User struct`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`ID    string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`json:"id"`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ModifyExistingStruct", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create initial struct
		err = project.Struct(gogo.StructOpts{
			Filename: "model.go",
			Name:     "Product",
			Fields: []gogo.StructField{
				{Name: "Name", Type: "string"},
				{Name: "Price", Type: "float64"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Modify the struct - add new field
		err = project.Struct(gogo.StructOpts{
			Filename: "model.go",
			Name:     "Product",
			Fields: []gogo.StructField{
				{Name: "Name", Type: "string"},
				{Name: "Price", Type: "float64"},
				{Name: "Stock", Type: "int"},
			},
			PreserveExisting: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify all fields exist
		if err := fs.Assert(`Name string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Price float64`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Stock int`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("StructWithContent", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "models",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Struct(gogo.StructOpts{
			Filename: "order.go",
			Name:     "Order",
			Content: `
				ID        string    ` + "`json:\"id\"`" + `
				Customer  string    ` + "`json:\"customer\"`" + `
				Items     []Item    ` + "`json:\"items\"`" + `
				Total     float64   ` + "`json:\"total\"`" + `
			`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify the struct was created with content
		if err := fs.Assert(`type Order struct`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Customer  string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Items     []Item`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DeleteFieldsFromStruct", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create initial struct with multiple fields
		err = project.Struct(gogo.StructOpts{
			Filename: "entity.go",
			Name:     "Entity",
			Fields: []gogo.StructField{
				{Name: "ID", Type: "int"},
				{Name: "Name", Type: "string"},
				{Name: "Deprecated", Type: "bool"},
				{Name: "Legacy", Type: "string"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Delete specific fields
		err = project.Struct(gogo.StructOpts{
			Filename: "entity.go",
			Name:     "Entity",
			Fields: []gogo.StructField{
				{Name: "ID", Type: "int"},
				{Name: "Name", Type: "string"},
			},
			DeleteFields: []gogo.StructField{
				{Name: "Deprecated"},
				{Name: "Legacy"},
			},
			PreserveExisting: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify fields were deleted
		if err := fs.Assert(`ID int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Name string`); err != nil {
			t.Fatal(err)
		}
		// These should not exist
		if err := fs.Assert(`Deprecated`); err == nil {
			t.Fatal("Deprecated field should have been deleted")
		}
		if err := fs.Assert(`Legacy`); err == nil {
			t.Fatal("Legacy field should have been deleted")
		}
	})

	t.Run("ComplexTypes", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Struct(gogo.StructOpts{
			Filename: "complex.go",
			Name:     "Complex",
			Fields: []gogo.StructField{
				{Name: "MapField", Type: "map[string]interface{}"},
				{Name: "SliceField", Type: "[]string"},
				{Name: "PointerField", Type: "*User"},
				{Name: "ChanField", Type: "chan int"},
				{Name: "FuncField", Type: "func(int) error"},
				{Name: "InterfaceField", Type: "interface{}"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify complex types are handled correctly
		if err := fs.Assert(`MapField map[string]interface{}`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`SliceField []string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`PointerField *User`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`ChanField chan int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`FuncField func(int) error`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`InterfaceField interface{}`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Missing filename
		err = project.Struct(gogo.StructOpts{
			Name: "Test",
			Fields: []gogo.StructField{
				{Name: "ID", Type: "int"},
			},
		})
		if err == nil {
			t.Fatal("Expected error for missing filename")
		}

		// Missing struct name
		err = project.Struct(gogo.StructOpts{
			Filename: "test.go",
			Fields: []gogo.StructField{
				{Name: "ID", Type: "int"},
			},
		})
		if err == nil {
			t.Fatal("Expected error for missing struct name")
		}

		// Both Fields and Content provided
		err = project.Struct(gogo.StructOpts{
			Filename: "test.go",
			Name:     "Test",
			Fields: []gogo.StructField{
				{Name: "ID", Type: "int"},
			},
			Content: "ID int",
		})
		if err == nil {
			t.Fatal("Expected error for mutually exclusive Fields and Content")
		}

		// Neither Fields nor Content provided
		err = project.Struct(gogo.StructOpts{
			Filename: "test.go",
			Name:     "Test",
		})
		if err == nil {
			t.Fatal("Expected error for missing Fields or Content")
		}
	})

	t.Run("MultipleStructsInSameFile", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "models",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add first struct
		err = project.Struct(gogo.StructOpts{
			Filename: "models.go",
			Name:     "User",
			Fields: []gogo.StructField{
				{Name: "ID", Type: "string"},
				{Name: "Name", Type: "string"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add second struct to same file
		err = project.Struct(gogo.StructOpts{
			Filename: "models.go",
			Name:     "Product",
			Fields: []gogo.StructField{
				{Name: "ID", Type: "string"},
				{Name: "Title", Type: "string"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify both structs exist
		if err := fs.Assert(`type User struct`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type Product struct`); err != nil {
			t.Fatal(err)
		}
	})
}
