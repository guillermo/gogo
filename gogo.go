package gogo

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"strings"
)

// OpenFS parses a directory containing Go files and returns a Template.
// It's a convenience wrapper around Open that uses the local filesystem.
// It reads all .go files in the specified directory and creates a Template
// with all the parsed files. The package name is determined from the first parsed file.
// It returns an error if the directory doesn't exist, isn't a directory, or if any files
// can't be parsed.
func OpenFS(dir string) (*Template, error) {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	// Create a filesystem from the directory
	fsys := os.DirFS(dir)

	// Use the more general Open function
	return Open(fsys, dir)
}

// Open parses Go files from a filesystem and returns a Template.
// It reads all .go files from the root of the provided filesystem and creates a Template
// with all the parsed files. The package name is determined from the first parsed file.
// The dirPath parameter is used to set the base directory path for the template.
// It returns an error if any files can't be parsed.
func Open(fsys fs.FS, dirPath string) (*Template, error) {
	fset := token.NewFileSet()
	template := &Template{
		Files: make(map[string]*File),
		dir:   dirPath,
		fset:  fset,
	}

	// Read all files from the filesystem
	files, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Process only .go files
	packageName := ""
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		// Open and read the file content
		f, err := fsys.Open(file.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", file.Name(), err)
		}

		// Parse the file content
		src, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file.Name(), err)
		}

		astFile, err := parser.ParseFile(fset, file.Name(), src, parser.ParseComments)
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
