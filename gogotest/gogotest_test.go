package gogotest

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("creates filesystem from text block", func(t *testing.T) {
		fs := New(`# main.go
package main

func main() {}

# user.go
package main

type User struct {
	ID string
}
`)

		files := fs.getFiles()
		if len(files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(files))
		}

		if !strings.Contains(string(files["main.go"]), "func main()") {
			t.Error("main.go should contain 'func main()'")
		}

		if !strings.Contains(string(files["user.go"]), "type User struct") {
			t.Error("user.go should contain 'type User struct'")
		}
	})

	t.Run("handles empty input", func(t *testing.T) {
		fs := New("")
		files := fs.getFiles()
		if len(files) != 0 {
			t.Errorf("Expected 0 files for empty input, got %d", len(files))
		}
	})

	t.Run("handles nested directories", func(t *testing.T) {
		fs := New(`# models/user.go
package models

type User struct {}

# controllers/user.go
package controllers
`)

		files := fs.getFiles()
		if len(files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(files))
		}

		if _, exists := files["models/user.go"]; !exists {
			t.Error("models/user.go should exist")
		}

		if _, exists := files["controllers/user.go"]; !exists {
			t.Error("controllers/user.go should exist")
		}
	})
}

func TestAssert(t *testing.T) {
	t.Run("finds content in files", func(t *testing.T) {
		fs := New(`# main.go
package main

type User struct {
	ID   string
	Name string
}
`)

		// Should find simple strings
		err := fs.Assert("package main")
		if err != nil {
			t.Errorf("Should find 'package main': %v", err)
		}

		// Should find struct definitions
		err = fs.Assert("type User struct")
		if err != nil {
			t.Errorf("Should find struct definition: %v", err)
		}

		// Should find fields
		err = fs.Assert("ID string")
		if err != nil {
			t.Errorf("Should find field: %v", err)
		}
	})

	t.Run("ignores whitespace differences", func(t *testing.T) {
		fs := New(`# test.go
package main

type Config struct {
	Host     string
	Port     int
	Debug    bool
}
`)

		// Should match despite different whitespace
		err := fs.Assert(`
type Config struct {
Host string
Port int
Debug bool
}`)
		if err != nil {
			t.Errorf("Should ignore whitespace differences: %v", err)
		}
	})

	t.Run("returns error when content not found", func(t *testing.T) {
		fs := New(`# main.go
package main
`)

		err := fs.Assert("nonexistent content")
		if err == nil {
			t.Error("Should return error for nonexistent content")
		}
	})
}

func TestString(t *testing.T) {
	t.Run("returns original format", func(t *testing.T) {
		input := `# main.go
package main

func main() {
	fmt.Println("Hello")
}

# user.go
package main

type User struct {
	ID string
}`

		fs := New(input)
		output := fs.String()

		// Should contain file markers
		if !strings.Contains(output, "# main.go") {
			t.Error("Output should contain '# main.go'")
		}

		if !strings.Contains(output, "# user.go") {
			t.Error("Output should contain '# user.go'")
		}

		// Should contain file contents
		if !strings.Contains(output, "func main()") {
			t.Error("Output should contain 'func main()'")
		}

		if !strings.Contains(output, "type User struct") {
			t.Error("Output should contain 'type User struct'")
		}
	})

	t.Run("returns empty string for empty filesystem", func(t *testing.T) {
		fs := New("")
		output := fs.String()
		if output != "" {
			t.Errorf("Expected empty string, got %q", output)
		}
	})
}
