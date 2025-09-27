package template_test

import (
	"fmt"
	"log"

	"github.com/guillermo/gogo"
	"github.com/guillermo/gogo/gogotest"
	"github.com/guillermo/gogo/template"
)

// Example demonstrates using the template package to transform a reference implementation
func Example() {
	// Create a reference implementation (e.g., an ORM with Customer model)
	referenceFS := gogotest.NewMemFS()
	referenceFS.WriteFile("customer.go", []byte(`package models

type Customer struct {
	CustomerID int
	Name       string
	Email      string
}

func GetCustomer(id int) *Customer {
	return &Customer{CustomerID: id}
}

func (c *Customer) Save() error {
	// Save customer
	return nil
}

const MaxCustomers = 1000
`), 0644)

	// Load the reference implementation as a template
	tmpl, err := template.New(referenceFS)
	if err != nil {
		log.Fatal(err)
	}

	// Transform: rename Customer to User
	tmpl, err = tmpl.RenameStruct("Customer", "User")
	if err != nil {
		log.Fatal(err)
	}

	// Transform: rename CustomerID field to UserID
	tmpl, err = tmpl.RenameStructField("User", "CustomerID", gogo.StructField{
		Name: "UserID",
		Type: "int",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Transform: rename function
	tmpl, err = tmpl.RenameFunction("GetCustomer", "GetUser")
	if err != nil {
		log.Fatal(err)
	}

	// Transform: rename constant
	tmpl, err = tmpl.RenameConstant("MaxCustomers", "MaxUsers")
	if err != nil {
		log.Fatal(err)
	}

	// Add a new field to the struct
	tmpl, err = tmpl.AddStructField("User", gogo.StructField{
		Name:       "CreatedAt",
		Type:       "time.Time",
		Annotation: "`json:\"created_at\"`",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Extract the transformed struct to use with gogo
	userStruct, err := tmpl.ExtractStruct("User")
	if err != nil {
		log.Fatal(err)
	}

	// Now use the transformed template with gogo to generate code in a target project
	targetFS := gogotest.NewMemFS()
	prj, err := gogo.New(gogo.Options{
		FS:                 targetFS,
		InitialPackageName: "models",
		ConflictFunc:       gogo.ConflictAccept, // Auto-accept for example
	})
	if err != nil {
		log.Fatal(err)
	}

	// Apply the extracted struct to the target project
	err = prj.Struct(gogo.StructOpts{
		Filename: "user.go",
		Name:     userStruct.Name,
		Fields:   userStruct.Fields,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully transformed Customer template to User model")
	// Output: Successfully transformed Customer template to User model
}
