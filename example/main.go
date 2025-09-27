package main

import (
	"fmt"
	"os"
	"time"

	"github.com/guillermo/gogo"
)

var dir = "/tmp/gogoexample"

func main() {
	fmt.Println("gogo Example 🚀")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Creating initial project...")

	createInitialProject()

	time.Sleep(500 * time.Millisecond)
	fmt.Println("Now, let's do some changes to the project")

	changeTheProject()
	time.Sleep(500 * time.Millisecond)

	fmt.Println("✅ Example completed! Check", dir, "for the generated files")
}

func createInitialProject() {
	// Clean up and create fresh directory
	os.RemoveAll(dir)
	os.MkdirAll(dir, 0755)

	filesystem, err := gogo.OpenFS(dir)
	if err != nil {
		panic(err)
	}

	project, err := gogo.New(gogo.Options{
		FS:                 filesystem,
		InitialPackageName: "main",
		ConflictFunc:       gogo.ConflictAccept,
	})
	if err != nil {
		panic(err)
	}

	err = project.Struct(gogo.StructOpts{
		Filename: "user.go",
		Name:     "User",
		Content: `
			ID   string ` + "`json:\"id\"`" + `
			Name string ` + "`json:\"name\"`" + `
		`,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("  ✅ Created user.go with User struct")
}

func changeTheProject() {
	filesystem, err := gogo.OpenFS(dir)
	if err != nil {
		panic(err)
	}

	project, err := gogo.New(gogo.Options{
		FS: filesystem,
	})
	if err != nil {
		panic(err)
	}

	// Add Age field to existing User struct
	err = project.Struct(gogo.StructOpts{
		Filename: "user.go",
		Name:     "User",
		Fields: []gogo.StructField{
			{Name: "ID", Type: "string", Annotation: `json:"id"`},
			{Name: "Name", Type: "string", Annotation: `json:"name"`},
			{Name: "Age", Type: "int", Annotation: `json:"age"`},
		},
		PreserveExisting: true,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("  ✅ Added Age field to User struct")
}
