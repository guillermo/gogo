package gogotest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/guillermo/gogo/fs"
)

// mockFileSystem implements a simple filesystem for testing
type mockFileSystem struct {
	mu    sync.RWMutex
	files map[string][]byte
	dirs  map[string]bool
}

// newMockFileSystem creates a new mock filesystem
func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

func (fs *mockFileSystem) writeFile(path string, data []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" && dir != "" {
		// Auto-create the parent directory for convenience
		fs.dirs[dir] = true
	}

	// Store a copy of the data
	fs.files[path] = append([]byte(nil), data...)
	return nil
}

func (fs *mockFileSystem) mkdirAll(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Mark all parent directories as created
	parts := strings.Split(path, string(filepath.Separator))
	current := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		if current == "" {
			current = part
		} else {
			current = filepath.Join(current, part)
		}
		fs.dirs[current] = true
	}

	return nil
}

// getFiles returns all files in the mock filesystem (for testing assertions)
func (fs *mockFileSystem) getFiles() map[string][]byte {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make(map[string][]byte)
	for k, v := range fs.files {
		result[k] = append([]byte(nil), v...)
	}
	return result
}

// mockFile implements a File interface for testing
type mockFile struct {
	io.ReadWriteCloser
	name    string
	content *bytes.Buffer
	closed  bool
}

func (f *mockFile) Read(p []byte) (n int, err error) {
	if f.closed {
		return 0, fmt.Errorf("file is closed")
	}
	return f.content.Read(p)
}

func (f *mockFile) Write(p []byte) (n int, err error) {
	if f.closed {
		return 0, fmt.Errorf("file is closed")
	}
	return f.content.Write(p)
}

func (f *mockFile) Close() error {
	f.closed = true
	return nil
}

func (f *mockFile) Name() string {
	return f.name
}

func (f *mockFile) Stat() (os.FileInfo, error) {
	return &mockFileInfo{
		name:    filepath.Base(f.name),
		size:    int64(f.content.Len()),
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}, nil
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *mockFileInfo) Name() string       { return fi.name }
func (fi *mockFileInfo) Size() int64        { return fi.size }
func (fi *mockFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *mockFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *mockFileInfo) IsDir() bool        { return fi.isDir }
func (fi *mockFileInfo) Sys() interface{}   { return nil }

// FS represents a mock filesystem for testing that implements gogo.FS interface
type FS struct {
	fs.FS
}

// gogo.FS interface methods
func (fs *mockFileSystem) ReadFile(path string) ([]byte, error) {
	files := fs.getFiles()
	content, exists := files[path]
	if !exists {
		return nil, os.ErrNotExist
	}
	return append([]byte(nil), content...), nil
}

func (fs *mockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return fs.writeFile(path, data)
}

func (fs *mockFileSystem) Stat(path string) (os.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Check if it's a directory
	if fs.dirs[path] {
		return &mockFileInfo{
			name:    filepath.Base(path),
			mode:    os.ModeDir | 0755,
			modTime: time.Now(),
			isDir:   true,
		}, nil
	}

	// Check if it's a file
	content, ok := fs.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}

	return &mockFileInfo{
		name:    filepath.Base(path),
		size:    int64(len(content)),
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}, nil
}

func (fs *mockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return fs.mkdirAll(path)
}

func (fs *mockFileSystem) Remove(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if it's a file
	if _, ok := fs.files[path]; ok {
		delete(fs.files, path)
		return nil
	}

	// Check if it's a directory
	if fs.dirs[path] {
		delete(fs.dirs, path)
		return nil
	}

	return os.ErrNotExist
}

func (fs *mockFileSystem) Rename(oldpath, newpath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if content, ok := fs.files[oldpath]; ok {
		fs.files[newpath] = content
		delete(fs.files, oldpath)
		return nil
	}

	return os.ErrNotExist
}

func (fs *mockFileSystem) TempFile(dir, pattern string) (fs.File, error) {
	name := fmt.Sprintf("%s/temp_%d.tmp", dir, time.Now().UnixNano())
	fs.writeFile(name, []byte{})

	return &mockFile{
		name:    name,
		content: bytes.NewBuffer([]byte{}),
	}, nil
}

func (fs *mockFileSystem) Open(path string) (fs.File, error) {
	content, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &mockFile{
		name:    path,
		content: bytes.NewBuffer(content),
	}, nil
}

func (fs *mockFileSystem) Create(path string) (fs.File, error) {
	fs.writeFile(path, []byte{})

	return &mockFile{
		name:    path,
		content: bytes.NewBuffer([]byte{}),
	}, nil
}

// GetFiles returns all files in the filesystem (for testing assertions)
func (fs *mockFileSystem) GetFiles() map[string][]byte {
	return fs.getFiles()
}

func (fs *FS) String() string {
	return fs.FS.(*mockFileSystem).String()
}

// String returns the content of the filesystem in the same text block format as input
func (fs *mockFileSystem) String() string {
	files := fs.getFiles()
	if len(files) == 0 {
		return ""
	}

	var result strings.Builder
	first := true

	for path, content := range files {
		if !first {
			result.WriteString("\n")
		}
		first = false

		result.WriteString("# ")
		result.WriteString(path)
		result.WriteString("\n")
		result.WriteString(string(content))
	}

	return result.String()
}

var _ fs.FS = new(mockFileSystem)

// Assert checks if the filesystem contains the expected content
// It ignores whitespace and indentation differences
func (fs *FS) Assert(expectedContent string) error {
	normalizedExpected := normalizeWhitespace(expectedContent)

	files := fs.FS.(*mockFileSystem).getFiles()
	for _, content := range files {
		normalizedContent := normalizeWhitespace(string(content))
		if strings.Contains(normalizedContent, normalizedExpected) {
			return nil // Found it!
		}
	}

	// Not found in any file
	var fileList []string
	for path := range files {
		fileList = append(fileList, path)
	}

	return fmt.Errorf("expected content not found in any file.\nExpected (normalized):\n%s\n\nSearched in files: %v",
		normalizedExpected, fileList)
}

// normalizeWhitespace removes extra whitespace and normalizes indentation
func normalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var normalized []string

	for _, line := range lines {
		// Trim all leading/trailing whitespace
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// Replace multiple spaces/tabs with single space
			trimmed = strings.ReplaceAll(trimmed, "\t", " ")
			for strings.Contains(trimmed, "  ") {
				trimmed = strings.ReplaceAll(trimmed, "  ", " ")
			}
			normalized = append(normalized, trimmed)
		}
	}

	return strings.Join(normalized, "\n")
}
