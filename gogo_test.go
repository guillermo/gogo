package gogo

import (
	"bytes"
	"go/ast"
	"os"
	"os/exec"
	"strings"
	"testing"
	"testing/fstest"
)

// fmt runs go fmt on the given directory
func gofmt(dir string) {
	cmd := exec.Command("go", "fmt")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func rundir(dir string) (out string, err error) {
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = dir
	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

func testHelloWorld(t *testing.T, outDir string, fn func(*Template)) {
	t.Helper()
	gofmt("fixtures/helloworld")
	os.RemoveAll(outDir)

	tmpl, err := OpenFS("fixtures/helloworld")
	if err != nil {
		t.Fatalf("failed to open fixture1: %v", err)
	}

	if fn != nil {
		fn(tmpl)
	}

	err = tmpl.Write(outDir)
	if err != nil {
		t.Fatal(err)
	}
	out, err := rundir(outDir)
	if err != nil {
		t.Fatalf("failed to run helloworld: %v", err)
	}
	if out != "HELLO, WORLD!" {
		t.Fatalf("expected Hello, World!, got %q", out)
	}

}

func TestJustCopy(t *testing.T) {
	testHelloWorld(t, "cases/helloworld", nil)
}

// TestOpenWithFSInterface tests if we can use the Open function with a custom fs.FS implementation
func TestOpenWithFSInterface(t *testing.T) {
	// Create an in-memory filesystem
	memFS := fstest.MapFS{
		"main.go": &fstest.MapFile{
			Data: []byte(`package main

import "fmt"

func main() {
	fmt.Println("HELLO, WORLD!")
}
`),
			Mode: 0644,
		},
	}

	// Open the filesystem
	tmpl, err := Open(memFS, "memory-fs")
	if err != nil {
		t.Fatalf("failed to open in-memory filesystem: %v", err)
	}

	// Check if the template was created correctly
	if tmpl.PackageName != "main" {
		t.Fatalf("expected package name 'main', got %q", tmpl.PackageName)
	}

	// Check if the file was read correctly
	file, ok := tmpl.Files["main.go"]
	if !ok {
		t.Fatal("main.go not found in template")
	}

	// Verify the file content by checking if the AST has the main function
	var mainFound bool
	for _, decl := range file.AstFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "main" {
			mainFound = true
			break
		}
	}

	if !mainFound {
		t.Fatal("main function not found in parsed AST")
	}

	// Verify the directory path was set correctly
	if tmpl.dir != "memory-fs" {
		t.Fatalf("expected dir to be 'memory-fs', got %q", tmpl.dir)
	}
}

func TestRenameFile(t *testing.T) {
	testHelloWorld(t, "cases/renamefile", func(tmpl *Template) {
		tmpl.RenameFile("helloworld.go", "main.go")
	})
}

func TestRenameType(t *testing.T) {
	testHelloWorld(t, "cases/renametype", func(tmpl *Template) {

		tmpl.OpenFile("helloworld.go", func(file *File) {
			err := file.RenameType("Message", "Output")
			if err != nil {
				t.Fatal(err)
			}
		})

	})

}

func TestRenameField(t *testing.T) {
	testHelloWorld(t, "cases/renamefield", func(tmpl *Template) {

		tmpl.OpenFile("helloworld.go", func(file *File) {
			err := file.OpenStruct("Message", func(s Struct) {
				err := s.Remove("Age")
				if err != nil {
					t.Fatal(err)
				}
			})
			if err != nil {
				t.Fatal(err)
			}
		})

	})

}

func TestAddField(t *testing.T) {
	testHelloWorld(t, "cases/addfield", func(tmpl *Template) {
		tmpl.OpenFile("helloworld.go", func(file *File) {
			err := file.OpenStruct("Message", func(s Struct) {
				s.Add("Priority", "int", map[string]string{"json": "age"})
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}

func TestDuplicateFile(t *testing.T) {
	testHelloWorld(t, "cases/duplicatefile", func(tmpl *Template) {
		file, err := tmpl.ExtractAndRemove("type.go")
		if err != nil {
			t.Fatal(err)
		}

		for _, nType := range []string{"Car", "Truck"} {
			clone := file.Clone()
			clone.RenameType("Type", nType)
			tmpl.Add(strings.ToLower(nType)+".go", clone)
		}

	})
}

func TestRemoveMethod(t *testing.T) {
	testHelloWorld(t, "cases/removemethod", func(tmpl *Template) {
		tmpl.OpenFile("helloworld.go", func(file *File) {
			err := file.OpenStruct("Message", func(s Struct) {
				err := s.RemoveMethod("Useless")
				if err != nil {
					t.Fatal(err)
				}
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}

func TestRenameMethod(t *testing.T) {
	testHelloWorld(t, "cases/renametype", func(tmpl *Template) {
		tmpl.OpenFile("helloworld.go", func(file *File) {
			err := file.OpenStruct("Message", func(s Struct) {
				err := s.RenameMethod("String", "Upper")
				if err != nil {
					t.Fatal(err)
				}
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}

func TestDuplicateMethod(t *testing.T) {
	testHelloWorld(t, "cases/duplicatemethod", func(tmpl *Template) {
		tmpl.OpenFile("helloworld.go", func(file *File) {
			err := file.OpenStruct("Message", func(s Struct) {
				err := s.DuplicateMethod("String", "Upper")
				if err != nil {
					t.Fatal(err)
				}
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}

func TestRemoveFunction(t *testing.T) {
	testHelloWorld(t, "cases/removefunction", func(tmpl *Template) {
		tmpl.OpenFile("helloworld.go", func(file *File) {
			err := file.RemoveFunction("main")
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}

func TestRenameFunction(t *testing.T) {
	testHelloWorld(t, "cases/renamefunction", func(tmpl *Template) {
		tmpl.OpenFile("helloworld.go", func(file *File) {
			err := file.RenameFunction("main", "main2")
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}

func TestDuplicateFunction(t *testing.T) {
	testHelloWorld(t, "cases/duplicatefunction", func(tmpl *Template) {
		err := tmpl.OpenFile("helloworld.go", func(file *File) {
			err := file.DuplicateFunction("main", "main2")
			if err != nil {
				t.Fatal(err)
			}
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPackageName(t *testing.T) {
	// It works but the testHelloWorld helper does not work as it is expecting the generated code to be in the main package
	//	testHelloWorld(t, "cases/packagename", func(tmpl *Template) {
	//		tmpl.PackageName = "main3"
	//	})
}

func run(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).Output()
	return string(out), err
}

func mustRun(cmd string) string {
	out, err := run(cmd)
	if err != nil {
		panic(err)
	}
	return out
}

func TestSave(t *testing.T) {
	outDir := "cases/save"
	gofmt("fixtures/existing")
	os.RemoveAll(outDir)
	mustRun("cp -r fixtures/existing " + outDir)

	tmpl, err := OpenFS(outDir)
	if err != nil {
		t.Fatalf("failed to open fixture1: %v", err)
	}

	err = tmpl.OpenStruct("User", func(s Struct) {
		s.Field("Name")
	})
	if err != nil {
		t.Fatal(err)
	}

	//
	out, err := run("diff -r fixtures/existing " + outDir)
	if err != nil {
		t.Fatal(out)
	}

	err = tmpl.Write(outDir)
	if err != nil {
		t.Fatal(err)
	}

}
