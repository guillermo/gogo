package gogo

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Open parses a directory containing Go files and returns a Template.
// It reads all .go files in the specified directory and creates a Template
// with all the parsed files. The package name is determined from the first parsed file.
// It returns an error if the directory doesn't exist, isn't a directory, or if any files
// can't be parsed.
func Open(dir string) (*Template, error) {
	fset := token.NewFileSet()
	template := &Template{
		Files: make(map[string]*File),
		dir:   dir,
		fset:  fset,
	}

	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	// Read all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Process only .go files
	packageName := ""
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		astFile, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file %s: %w", file.Name(), err)
		}

		// Set package name from the first parsed file
		if packageName == "" {
			packageName = astFile.Name.Name
		}

		template.Files[file.Name()] = &File{
			Name:     file.Name(),
			AstFile:  astFile,
			Template: template,
			fset:     fset,
		}
	}

	template.PackageName = packageName
	return template, nil
}

// ErrNotFound is returned when an item is not found in the codebase.
var ErrNotFound = errors.New("not found")
