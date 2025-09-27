package gogotest

import (
	"path/filepath"
	"strings"
)

// New creates a mock filesystem from a text block format
// The format is:
// # path/to/file.go
// file contents
//
// # another/file.go
// more contents
func New(content string) *FS {
	mockFS := newMockFileSystem()
	fs := &FS{
		FS: mockFS,
	}

	// Split by lines
	lines := strings.Split(content, "\n")

	var currentFile string
	var fileContent strings.Builder

	for _, line := range lines {
		// Check if this is a file marker
		if strings.HasPrefix(line, "# ") {
			// Save previous file if exists
			if currentFile != "" {
				// Ensure directory exists
				dir := filepath.Dir(currentFile)
				if dir != "." && dir != "/" {
					fs.FS.(*mockFileSystem).mkdirAll(dir)
				}
				// Write file
				fs.FS.(*mockFileSystem).writeFile(currentFile, []byte(fileContent.String()))
			}

			// Start new file
			currentFile = strings.TrimPrefix(line, "# ")
			currentFile = strings.TrimSpace(currentFile)
			fileContent.Reset()
		} else {
			// Add to current file content
			if fileContent.Len() > 0 {
				fileContent.WriteString("\n")
			}
			fileContent.WriteString(line)
		}
	}

	// Save last file if exists
	if currentFile != "" {
		// Ensure directory exists
		dir := filepath.Dir(currentFile)
		if dir != "." && dir != "/" {
			fs.FS.(*mockFileSystem).mkdirAll(dir)
		}
		// Write file
		fs.FS.(*mockFileSystem).writeFile(currentFile, []byte(fileContent.String()))
	}

	return fs
}

// parseFiles parses the text block format into a map of path -> content
// This is useful for the SetupTestProject function
func parseFiles(content string) map[string]string {
	files := make(map[string]string)

	// Split by lines
	lines := strings.Split(content, "\n")

	var currentFile string
	var fileContent strings.Builder

	for _, line := range lines {
		// Check if this is a file marker
		if strings.HasPrefix(line, "# ") {
			// Save previous file if exists
			if currentFile != "" {
				files[currentFile] = fileContent.String()
			}

			// Start new file
			currentFile = strings.TrimPrefix(line, "# ")
			currentFile = strings.TrimSpace(currentFile)
			fileContent.Reset()
		} else {
			// Add to current file content
			if fileContent.Len() > 0 {
				fileContent.WriteString("\n")
			}
			fileContent.WriteString(line)
		}
	}

	// Save last file if exists
	if currentFile != "" {
		files[currentFile] = fileContent.String()
	}

	return files
}

// NewMemFS creates an empty mock filesystem for testing
func NewMemFS() *mockFileSystem {
	return newMockFileSystem()
}
