package realfs

import (
	"io"
	"os"

	"github.com/guillermo/gogo/fs"
)

// File interface (matches gogo.File)
type File interface {
	io.ReadWriteCloser
	Name() string
	Stat() (os.FileInfo, error)
}

// FileWrapper wraps os.File to implement the File interface
type FileWrapper struct {
	*os.File
}

// Implement the required interface methods
func (f *FileWrapper) Read(p []byte) (n int, err error) {
	return f.File.Read(p)
}

func (f *FileWrapper) Write(p []byte) (n int, err error) {
	return f.File.Write(p)
}

func (f *FileWrapper) Close() error {
	return f.File.Close()
}

func (f *FileWrapper) Name() string {
	return f.File.Name()
}

func (f *FileWrapper) Stat() (os.FileInfo, error) {
	return f.File.Stat()
}

// FS implements FileSystem using actual OS operations
type FS struct{}

func (fs *FS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (fs *FS) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (fs *FS) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (fs *FS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs *FS) Remove(path string) error {
	return os.Remove(path)
}

func (fs *FS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (fs *FS) TempFile(dir, pattern string) (fs.File, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, err
	}
	return &FileWrapper{f}, nil
}

func (fs *FS) Open(path string) (fs.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &FileWrapper{f}, nil
}

func (fs *FS) Create(path string) (fs.File, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &FileWrapper{f}, nil
}

// Open creates a new filesystem instance for the given path.
// If the directory doesn't exist, it will be created recursively.
func Open(path string) (fs.FS, error) {
	// Check if the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create the directory recursively
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, err
		}
	} else if err != nil {
		// Some other error occurred while checking the path
		return nil, err
	}

	// Return a new filesystem instance
	return &FS{}, nil
}
