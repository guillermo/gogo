package template

import (
	"testing"

	"github.com/guillermo/gogo"
	"github.com/guillermo/gogo/gogotest"
)

func TestNew(t *testing.T) {
	// Create a test filesystem with a simple Go file
	fs := gogotest.NewMemFS()
	fs.WriteFile("customer.go", []byte(`package models

type Customer struct {
	ID   int
	Name string
}

func GetCustomer(id int) *Customer {
	return &Customer{ID: id}
}

var DefaultCustomer = &Customer{ID: 1, Name: "Default"}

const MaxCustomers = 100
`), 0644)

	// Create a template from the filesystem
	tmpl, err := New(fs)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	if tmpl == nil {
		t.Fatal("Template is nil")
	}

	if tmpl.pkgName != "models" {
		t.Errorf("Expected package name 'models', got '%s'", tmpl.pkgName)
	}
}

func TestExtractStruct(t *testing.T) {
	fs := gogotest.NewMemFS()
	fs.WriteFile("customer.go", []byte(`package models

type Customer struct {
	ID   int    `+"`json:\"id\"`"+`
	Name string `+"`json:\"name\"`"+`
}
`), 0644)

	tmpl, err := New(fs)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Extract the Customer struct
	structOpts, err := tmpl.ExtractStruct("Customer")
	if err != nil {
		t.Fatalf("Failed to extract struct: %v", err)
	}

	if structOpts.Name != "Customer" {
		t.Errorf("Expected struct name 'Customer', got '%s'", structOpts.Name)
	}

	if len(structOpts.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(structOpts.Fields))
	}

	if structOpts.Fields[0].Name != "ID" {
		t.Errorf("Expected first field name 'ID', got '%s'", structOpts.Fields[0].Name)
	}

	if structOpts.Fields[0].Type != "int" {
		t.Errorf("Expected first field type 'int', got '%s'", structOpts.Fields[0].Type)
	}
}

func TestRenameStruct(t *testing.T) {
	fs := gogotest.NewMemFS()
	fs.WriteFile("customer.go", []byte(`package models

type Customer struct {
	ID   int
	Name string
}

func GetCustomer(id int) *Customer {
	return &Customer{ID: id}
}
`), 0644)

	tmpl, err := New(fs)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Rename Customer to User
	newTmpl, err := tmpl.RenameStruct("Customer", "User")
	if err != nil {
		t.Fatalf("Failed to rename struct: %v", err)
	}

	// Verify the struct was renamed
	_, err = newTmpl.ExtractStruct("Customer")
	if err == nil {
		t.Error("Expected error when extracting old struct name")
	}

	userOpts, err := newTmpl.ExtractStruct("User")
	if err != nil {
		t.Fatalf("Failed to extract renamed struct: %v", err)
	}

	if userOpts.Name != "User" {
		t.Errorf("Expected struct name 'User', got '%s'", userOpts.Name)
	}
}

func TestRenameStructField(t *testing.T) {
	fs := gogotest.NewMemFS()
	fs.WriteFile("customer.go", []byte(`package models

type Customer struct {
	CustomerID int
	Name       string
}
`), 0644)

	tmpl, err := New(fs)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Rename CustomerID to UserID
	newTmpl, err := tmpl.RenameStructField("Customer", "CustomerID", gogo.StructField{
		Name: "UserID",
		Type: "int",
	})
	if err != nil {
		t.Fatalf("Failed to rename struct field: %v", err)
	}

	// Verify the field was renamed
	customerOpts, err := newTmpl.ExtractStruct("Customer")
	if err != nil {
		t.Fatalf("Failed to extract struct: %v", err)
	}

	foundNewField := false
	for _, field := range customerOpts.Fields {
		if field.Name == "CustomerID" {
			t.Error("Old field name still exists")
		}
		if field.Name == "UserID" {
			foundNewField = true
		}
	}

	if !foundNewField {
		t.Error("New field name not found")
	}
}

func TestAddStructField(t *testing.T) {
	fs := gogotest.NewMemFS()
	fs.WriteFile("customer.go", []byte(`package models

type Customer struct {
	ID   int
	Name string
}
`), 0644)

	tmpl, err := New(fs)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Add a new field
	newField := gogo.StructField{Name: "Email", Type: "string", Annotation: "`json:\"email\"`"}

	newTmpl, err := tmpl.AddStructField("Customer", newField)
	if err != nil {
		t.Fatalf("Failed to add struct field: %v", err)
	}

	// Verify the field was added
	customerOpts, err := newTmpl.ExtractStruct("Customer")
	if err != nil {
		t.Fatalf("Failed to extract struct: %v", err)
	}

	if len(customerOpts.Fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(customerOpts.Fields))
	}

	foundNewField := false
	for _, field := range customerOpts.Fields {
		if field.Name == "Email" {
			foundNewField = true
			if field.Type != "string" {
				t.Errorf("Expected field type 'string', got '%s'", field.Type)
			}
		}
	}

	if !foundNewField {
		t.Error("New field not found")
	}
}

func TestRemoveStructField(t *testing.T) {
	fs := gogotest.NewMemFS()
	fs.WriteFile("customer.go", []byte(`package models

type Customer struct {
	ID    int
	Name  string
	Email string
}
`), 0644)

	tmpl, err := New(fs)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Remove the Email field
	fieldToRemove := gogo.StructField{Name: "Email"}

	newTmpl, err := tmpl.RemoveStructField("Customer", fieldToRemove)
	if err != nil {
		t.Fatalf("Failed to remove struct field: %v", err)
	}

	// Verify the field was removed
	customerOpts, err := newTmpl.ExtractStruct("Customer")
	if err != nil {
		t.Fatalf("Failed to extract struct: %v", err)
	}

	if len(customerOpts.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(customerOpts.Fields))
	}

	for _, field := range customerOpts.Fields {
		if field.Name == "Email" {
			t.Error("Email field should have been removed")
		}
	}
}

func TestRenameVariable(t *testing.T) {
	fs := gogotest.NewMemFS()
	fs.WriteFile("customer.go", []byte(`package models

var DefaultCustomer = &Customer{ID: 1}
`), 0644)

	tmpl, err := New(fs)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Rename variable
	newTmpl, err := tmpl.RenameVariable("DefaultCustomer", "DefaultUser")
	if err != nil {
		t.Fatalf("Failed to rename variable: %v", err)
	}

	// Verify the variable was renamed
	_, err = newTmpl.ExtractVariable("DefaultCustomer")
	if err == nil {
		t.Error("Expected error when extracting old variable name")
	}

	_, err = newTmpl.ExtractVariable("DefaultUser")
	if err != nil {
		t.Errorf("Failed to extract renamed variable: %v", err)
	}
}

func TestRenameFunction(t *testing.T) {
	fs := gogotest.NewMemFS()
	fs.WriteFile("customer.go", []byte(`package models

type Customer struct {
	ID int
}

func GetCustomer(id int) *Customer {
	return &Customer{ID: id}
}
`), 0644)

	tmpl, err := New(fs)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Rename function
	newTmpl, err := tmpl.RenameFunction("GetCustomer", "GetUser")
	if err != nil {
		t.Fatalf("Failed to rename function: %v", err)
	}

	// Verify the function was renamed
	_, err = newTmpl.ExtractFunction("GetCustomer")
	if err == nil {
		t.Error("Expected error when extracting old function name")
	}

	funcOpts, err := newTmpl.ExtractFunction("GetUser")
	if err != nil {
		t.Fatalf("Failed to extract renamed function: %v", err)
	}

	if funcOpts.Name != "GetUser" {
		t.Errorf("Expected function name 'GetUser', got '%s'", funcOpts.Name)
	}
}

func TestRenameConstant(t *testing.T) {
	fs := gogotest.NewMemFS()
	fs.WriteFile("customer.go", []byte(`package models

const MaxCustomers = 100
`), 0644)

	tmpl, err := New(fs)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Rename constant
	newTmpl, err := tmpl.RenameConstant("MaxCustomers", "MaxUsers")
	if err != nil {
		t.Fatalf("Failed to rename constant: %v", err)
	}

	// Verify the constant was renamed
	_, err = newTmpl.ExtractConstant("MaxCustomers")
	if err == nil {
		t.Error("Expected error when extracting old constant name")
	}

	_, err = newTmpl.ExtractConstant("MaxUsers")
	if err != nil {
		t.Errorf("Failed to extract renamed constant: %v", err)
	}
}

func TestChainedTransformations(t *testing.T) {
	fs := gogotest.NewMemFS()
	fs.WriteFile("customer.go", []byte(`package models

type Customer struct {
	CustomerID int
	Name       string
}

func GetCustomer(id int) *Customer {
	return &Customer{CustomerID: id}
}

const MaxCustomers = 100
`), 0644)

	tmpl, err := New(fs)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Chain multiple transformations
	tmpl, err = tmpl.RenameStruct("Customer", "User")
	if err != nil {
		t.Fatalf("Failed to rename struct: %v", err)
	}

	tmpl, err = tmpl.RenameStructField("User", "CustomerID", gogo.StructField{
		Name: "UserID",
		Type: "int",
	})
	if err != nil {
		t.Fatalf("Failed to rename field: %v", err)
	}

	tmpl, err = tmpl.RenameFunction("GetCustomer", "GetUser")
	if err != nil {
		t.Fatalf("Failed to rename function: %v", err)
	}

	tmpl, err = tmpl.RenameConstant("MaxCustomers", "MaxUsers")
	if err != nil {
		t.Fatalf("Failed to rename constant: %v", err)
	}

	// Verify all transformations were applied
	userOpts, err := tmpl.ExtractStruct("User")
	if err != nil {
		t.Fatalf("Failed to extract User struct: %v", err)
	}

	if userOpts.Name != "User" {
		t.Errorf("Expected struct name 'User', got '%s'", userOpts.Name)
	}

	foundUserID := false
	for _, field := range userOpts.Fields {
		if field.Name == "UserID" {
			foundUserID = true
		}
		if field.Name == "CustomerID" {
			t.Error("Old field name CustomerID still exists")
		}
	}

	if !foundUserID {
		t.Error("UserID field not found")
	}
}
