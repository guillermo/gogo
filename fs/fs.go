package fs

import (
	"io"
	"os"
)

// FS defines the interface for file operations
type FS interface {
	// ReadFile reads the entire file content
	ReadFile(path string) ([]byte, error)

	// WriteFile writes data to a file
	WriteFile(path string, data []byte, perm os.FileMode) error

	// Stat returns file info
	Stat(path string) (os.FileInfo, error)

	// MkdirAll creates a directory path
	MkdirAll(path string, perm os.FileMode) error

	// Remove removes a file or empty directory
	Remove(path string) error

	// Rename renames/moves a file
	Rename(oldpath, newpath string) error

	// TempFile creates a temporary file
	TempFile(dir, pattern string) (File, error)

	// Open opens a file for reading
	Open(path string) (File, error)

	// Create creates or truncates a file
	Create(path string) (File, error)
}

// File represents an open file
type File interface {
	io.ReadWriteCloser
	Name() string
	Stat() (os.FileInfo, error)
}
