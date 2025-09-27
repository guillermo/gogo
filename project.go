package gogo

import (
	"fmt"
	"path/filepath"

	"github.com/guillermo/gogo/fs"
)

// Project represents a Go project being modified
type Project struct {
	opts         Options
	fs           fs.FS
	conflictFunc ConflictFunc
}

// modifyStructInFile modifies or creates a struct in Go source code
func (p *Project) modifyStructInFile(content []byte, s structDef, defaultPackage string) ([]byte, error) {
	// This will be implemented with AST parsing
	// For now, return a placeholder
	if len(content) == 0 {
		// Create new file
		return createNewFileWithStruct(s, defaultPackage)
	}

	// Modify existing file
	return modifyExistingFile(content, s)
}

// modifyMethodInFile modifies or creates methods in Go source code
func (p *Project) modifyMethodInFile(content []byte, opts MethodOpts, defaultPackage string) ([]byte, error) {
	// This will be implemented with AST parsing
	// For now, return a placeholder
	if len(content) == 0 {
		// Create new file
		return createNewFileWithMethod(opts, defaultPackage)
	}

	// Modify existing file
	return modifyExistingFileForMethod(content, opts)
}

// modifyFunctionInFile modifies or creates functions in Go source code
func (p *Project) modifyFunctionInFile(content []byte, opts FunctionOpts, defaultPackage string) ([]byte, error) {
	// This will be implemented with AST parsing
	// For now, return a placeholder
	if len(content) == 0 {
		// Create new file
		return createNewFileWithFunction(opts, defaultPackage)
	}

	// Modify existing file
	return modifyExistingFileForFunction(content, opts)
}

// modifyVariableInFile modifies or creates variables in Go source code
func (p *Project) modifyVariableInFile(content []byte, opts VariableOpts, defaultPackage string) ([]byte, error) {
	// This will be implemented with AST parsing
	// For now, return a placeholder
	if len(content) == 0 {
		// Create new file
		return createNewFileWithVariable(opts, defaultPackage)
	}

	// Modify existing file
	return modifyExistingFileForVariable(content, opts)
}

// modifyConstantInFile modifies or creates constants in Go source code
func (p *Project) modifyConstantInFile(content []byte, opts ConstantOpts, defaultPackage string) ([]byte, error) {
	// This will be implemented with AST parsing
	// For now, return a placeholder
	if len(content) == 0 {
		// Create new file
		return createNewFileWithConstant(opts, defaultPackage)
	}

	// Modify existing file
	return modifyExistingFileForConstant(content, opts)
}

// modifyTypeInFile modifies or creates type definitions in Go source code
func (p *Project) modifyTypeInFile(content []byte, opts TypeOpts, defaultPackage string) ([]byte, error) {
	// This will be implemented with AST parsing
	// For now, return a placeholder
	if len(content) == 0 {
		// Create new file
		return createNewFileWithType(opts, defaultPackage)
	}

	// Modify existing file
	return modifyExistingFileForType(content, opts)
}

// Struct creates or modifies a struct using the unified API
func (p *Project) Struct(opts StructOpts) error {
	// Validation: Fields and Content are mutually exclusive
	if len(opts.Fields) > 0 && opts.Content != "" {
		return fmt.Errorf("Fields and Content are mutually exclusive - provide only one")
	}

	// Validation: Must provide either Fields or Content
	if len(opts.Fields) == 0 && opts.Content == "" {
		return fmt.Errorf("must provide either Fields or Content")
	}

	// Validation: Required fields
	if opts.Filename == "" {
		return fmt.Errorf("Filename is required")
	}
	if opts.Name == "" {
		return fmt.Errorf("struct Name is required")
	}

	// Prepare the struct definition
	var s structDef
	if opts.Content != "" {
		// Parse the content string to create StructField slice (like CreateOrSetMethodS)
		parsedFields, err := parseFieldsString(opts.Content)
		if err != nil {
			return fmt.Errorf("failed to parse fields: %w", err)
		}
		s = structDef{
			Name:             opts.Name,
			EnsureFields:     parsedFields,
			PreserveExisting: true,
		}
	} else {
		// Use Fields-based approach (like CreateOrSetStruct)
		s = structDef{
			Name:             opts.Name,
			EnsureFields:     opts.Fields,
			DeleteFields:     opts.DeleteFields,
			PreserveExisting: opts.PreserveExisting,
		}
	}

	// Read existing file content if it exists
	var oldContent []byte
	_, err := p.fs.Stat(opts.Filename)
	fileExists := err == nil

	if fileExists {
		oldContent, err = p.fs.ReadFile(opts.Filename)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
	}

	// Parse and modify the content
	newContent, err := p.modifyStructInFile(oldContent, s, p.opts.InitialPackageName)
	if err != nil {
		return fmt.Errorf("failed to modify struct: %w", err)
	}

	// Check if there are actual changes
	if fileExists && string(oldContent) == string(newContent) {
		// No changes needed
		return nil
	}

	// Apply the changes
	return p.applyChanges(opts.Filename, oldContent, newContent, fileExists)
}

// Method creates or modifies a method using the unified API
func (p *Project) Method(opts MethodOpts) error {
	// Validation: Parameters/ReturnType/Body and Content are mutually exclusive
	hasStructuredParams := len(opts.Parameters) > 0 || opts.ReturnType != "" || opts.Body != ""
	if hasStructuredParams && opts.Content != "" {
		return fmt.Errorf("Parameters/ReturnType/Body and Content are mutually exclusive - provide only one approach")
	}

	// Validation: Must provide either structured params or content
	if !hasStructuredParams && opts.Content == "" {
		return fmt.Errorf("must provide either Parameters/ReturnType/Body or Content")
	}

	// Validation: Required fields
	if opts.Filename == "" {
		return fmt.Errorf("Filename is required")
	}
	if opts.Name == "" {
		return fmt.Errorf("method Name is required")
	}
	if opts.ReceiverType == "" {
		return fmt.Errorf("ReceiverType is required for methods")
	}

	// Read existing file content if it exists
	var oldContent []byte
	_, err := p.fs.Stat(opts.Filename)
	fileExists := err == nil

	if fileExists {
		oldContent, err = p.fs.ReadFile(opts.Filename)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
	}

	// Parse and modify the content
	newContent, err := p.modifyMethodInFile(oldContent, opts, p.opts.InitialPackageName)
	if err != nil {
		return fmt.Errorf("failed to modify method: %w", err)
	}

	// Check if there are actual changes
	if fileExists && string(oldContent) == string(newContent) {
		// No changes needed
		return nil
	}

	// Apply the changes using the same pattern as Struct
	return p.applyChanges(opts.Filename, oldContent, newContent, fileExists)
}

// Function creates or modifies a function using the unified API
func (p *Project) Function(opts FunctionOpts) error {
	// Validation: Parameters/ReturnType/Body and Content are mutually exclusive
	hasStructuredParams := len(opts.Parameters) > 0 || opts.ReturnType != "" || opts.Body != ""
	if hasStructuredParams && opts.Content != "" {
		return fmt.Errorf("Parameters/ReturnType/Body and Content are mutually exclusive - provide only one approach")
	}

	// Validation: Must provide either structured params or content
	if !hasStructuredParams && opts.Content == "" {
		return fmt.Errorf("must provide either Parameters/ReturnType/Body or Content")
	}

	// Validation: Required fields
	if opts.Filename == "" {
		return fmt.Errorf("Filename is required")
	}
	if opts.Name == "" {
		return fmt.Errorf("function Name is required")
	}

	// Read existing file content if it exists
	var oldContent []byte
	_, err := p.fs.Stat(opts.Filename)
	fileExists := err == nil

	if fileExists {
		oldContent, err = p.fs.ReadFile(opts.Filename)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
	}

	// Parse and modify the content
	newContent, err := p.modifyFunctionInFile(oldContent, opts, p.opts.InitialPackageName)
	if err != nil {
		return fmt.Errorf("failed to modify function: %w", err)
	}

	// Check if there are actual changes
	if fileExists && string(oldContent) == string(newContent) {
		// No changes needed
		return nil
	}

	// Apply the changes
	return p.applyChanges(opts.Filename, oldContent, newContent, fileExists)
}

// Variable creates or modifies variables using the unified API
func (p *Project) Variable(opts VariableOpts) error {
	// Validation: Variables and Content are mutually exclusive
	if len(opts.Variables) > 0 && opts.Content != "" {
		return fmt.Errorf("Variables and Content are mutually exclusive - provide only one")
	}

	// Validation: Must provide either Variables or Content
	if len(opts.Variables) == 0 && opts.Content == "" {
		return fmt.Errorf("must provide either Variables or Content")
	}

	// Validation: Required fields
	if opts.Filename == "" {
		return fmt.Errorf("Filename is required")
	}

	// Read existing file content if it exists
	var oldContent []byte
	_, err := p.fs.Stat(opts.Filename)
	fileExists := err == nil

	if fileExists {
		oldContent, err = p.fs.ReadFile(opts.Filename)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
	}

	// Parse and modify the content
	newContent, err := p.modifyVariableInFile(oldContent, opts, p.opts.InitialPackageName)
	if err != nil {
		return fmt.Errorf("failed to modify variables: %w", err)
	}

	// Check if there are actual changes
	if fileExists && string(oldContent) == string(newContent) {
		// No changes needed
		return nil
	}

	// Apply the changes
	return p.applyChanges(opts.Filename, oldContent, newContent, fileExists)
}

// Constant creates or modifies constants using the unified API
func (p *Project) Constant(opts ConstantOpts) error {
	// Validation: Constants and Content are mutually exclusive
	if len(opts.Constants) > 0 && opts.Content != "" {
		return fmt.Errorf("Constants and Content are mutually exclusive - provide only one")
	}

	// Validation: Must provide either Constants or Content
	if len(opts.Constants) == 0 && opts.Content == "" {
		return fmt.Errorf("must provide either Constants or Content")
	}

	// Validation: Required fields
	if opts.Filename == "" {
		return fmt.Errorf("Filename is required")
	}

	// Read existing file content if it exists
	var oldContent []byte
	_, err := p.fs.Stat(opts.Filename)
	fileExists := err == nil

	if fileExists {
		oldContent, err = p.fs.ReadFile(opts.Filename)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
	}

	// Parse and modify the content
	newContent, err := p.modifyConstantInFile(oldContent, opts, p.opts.InitialPackageName)
	if err != nil {
		return fmt.Errorf("failed to modify constants: %w", err)
	}

	// Check if there are actual changes
	if fileExists && string(oldContent) == string(newContent) {
		// No changes needed
		return nil
	}

	// Apply the changes
	return p.applyChanges(opts.Filename, oldContent, newContent, fileExists)
}

// Type creates or modifies type definitions using the unified API
func (p *Project) Type(opts TypeOpts) error {
	// Validation: Types and Content are mutually exclusive
	if len(opts.Types) > 0 && opts.Content != "" {
		return fmt.Errorf("Types and Content are mutually exclusive - provide only one")
	}

	// Validation: Must provide either Types or Content
	if len(opts.Types) == 0 && opts.Content == "" {
		return fmt.Errorf("must provide either Types or Content")
	}

	// Validation: Required fields
	if opts.Filename == "" {
		return fmt.Errorf("Filename is required")
	}

	// Read existing file content if it exists
	var oldContent []byte
	_, err := p.fs.Stat(opts.Filename)
	fileExists := err == nil

	if fileExists {
		oldContent, err = p.fs.ReadFile(opts.Filename)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
	}

	// Parse and modify the content
	newContent, err := p.modifyTypeInFile(oldContent, opts, p.opts.InitialPackageName)
	if err != nil {
		return fmt.Errorf("failed to modify types: %w", err)
	}

	// Check if there are actual changes
	if fileExists && string(oldContent) == string(newContent) {
		// No changes needed
		return nil
	}

	// Apply the changes
	return p.applyChanges(opts.Filename, oldContent, newContent, fileExists)
}

// applyChanges applies the changes to a file using the common pattern
func (p *Project) applyChanges(filename string, oldContent, newContent []byte, fileExists bool) error {
	// Ensure the directory exists
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		if err := p.fs.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create temp file
	tempFile, err := p.fs.TempFile(dir, ".gogo-*.go")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()

	// Write new content to temp file
	if err := p.fs.WriteFile(tempPath, newContent, 0644); err != nil {
		p.fs.Remove(tempPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Prepare change info
	action := "modify"
	if !fileExists {
		action = "create"
	}

	changeInfo := ChangeInfo{
		Action:     action,
		FileName:   filename,
		OldContent: oldContent,
		NewContent: newContent,
		Diff:       generateDiff(oldContent, newContent, filename),
	}

	// Ask for confirmation if needed
	if p.conflictFunc != nil {
		if !p.conflictFunc(p.fs, filename, tempPath, changeInfo) {
			// User rejected changes
			p.fs.Remove(tempPath)
			return nil
		}
	}

	// Apply changes by moving temp file to target
	if fileExists {
		// Backup existing file temporarily
		backupPath := filename + ".backup"
		if err := p.fs.Rename(filename, backupPath); err != nil {
			p.fs.Remove(tempPath)
			return fmt.Errorf("failed to backup existing file: %w", err)
		}

		// Move temp file to target
		if err := p.fs.Rename(tempPath, filename); err != nil {
			// Restore backup
			p.fs.Rename(backupPath, filename)
			p.fs.Remove(tempPath)
			return fmt.Errorf("failed to apply changes: %w", err)
		}

		// Remove backup
		p.fs.Remove(backupPath)
	} else {
		// Just move temp file to target
		if err := p.fs.Rename(tempPath, filename); err != nil {
			p.fs.Remove(tempPath)
			return fmt.Errorf("failed to create file: %w", err)
		}
	}

	return nil
}
