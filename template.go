package gogo

import (
	"fmt"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
)

// Template represents a Go package template that can be manipulated.
// It contains a collection of Go files that can be modified and written to disk.
type Template struct {
	PackageName string           // PackageName is the name of the package
	Files       map[string]*File // Files is a map of filenames to File objects
	dir         string           // dir is the directory where the template was loaded from
	fset        *token.FileSet   // fset is the file set used for position information
}

func (t *Template) OpenStruct(name string, modifier StructModifier) error {
	for _, file := range t.Files {
		err := file.OpenStruct(name, modifier)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			panic(err)
		}
		return nil
	}
	return ErrNotFound
}

// Write writes the template to a directory.
// It creates the directory if it doesn't exist and formats all Go files
// with the updated package name before writing them to disk.
func (t *Template) Write(dir string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write each file
	for name, file := range t.Files {
		// Update package name
		file.AstFile.Name.Name = t.PackageName

		// Format the AST
		filePath := filepath.Join(dir, name)
		f, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", name, err)
		}

		if err := format.Node(f, t.fset, file.AstFile); err != nil {
			f.Close()
			return fmt.Errorf("failed to format file %s: %w", name, err)
		}

		f.Close()
	}

	return nil
}

// ExtractAndRemove extracts a file from the template and removes it.
// It returns a copy of the file and removes the original from the template.
// If the file doesn't exist, it returns an error.
func (t *Template) ExtractAndRemove(name string) (*File, error) {
	file, ok := t.Files[name]
	if !ok {
		return nil, fmt.Errorf("file %s not found", name)
	}

	// Create a copy of the file
	fileCopy := &File{
		Name:     file.Name,
		AstFile:  file.AstFile,
		Template: file.Template,
		fset:     file.fset,
	}

	// Remove the file from the template
	delete(t.Files, name)

	return fileCopy, nil
}

// Add adds a file to the template.
// It updates the file's name and adds it to the template's Files map.
func (t *Template) Add(name string, file *File) {
	// Update file name
	file.Name = name
	t.Files[name] = file
}

// OpenFile opens a file from the template and applies a modifier function to it.
// It returns an error if the file doesn't exist.
func (t *Template) OpenFile(name string, modifier func(*File)) error {
	file, ok := t.Files[name]
	if !ok {
		return fmt.Errorf("file %s not found", name)
	}

	modifier(file)

	return nil
}

// RenameFile renames a file in the template.
// It extracts the file with the old name and adds it back with the new name.
// If the file with the old name doesn't exist, it returns an error.
func (t *Template) RenameFile(oldName, newName string) error {
	f, err := t.ExtractAndRemove(oldName)
	if err != nil {
		return err
	}
	t.Add(newName, f)
	return nil
}
