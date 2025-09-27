// Package gogo provides a library for programmatically generating and modifying Go source files.
//
// It's designed to create Go code from within Go programs, enabling use cases like:
//   - Building ORMs and code generators
//   - Scaffolding tools
//   - Automated refactoring
//   - Template-based code generation
//
// The library provides a simplified API focused on direct code manipulation without source templates.
//
// Example:
//
//	prj, _ := gogo.NewFS("./target", gogo.Options{
//	    InitialPackageName: "models",
//	    ConflictFunc:       gogo.ConflictAsk,
//	})
//
//	prj.Struct(gogo.StructOpts{
//	    Filename: "user.go",
//	    Name:     "User",
//	    Fields: []gogo.StructField{
//	        {Name: "ID", Type: "int"},
//	        {Name: "Name", Type: "string"},
//	    },
//	})
package gogo

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/guillermo/gogo/fs"
	"github.com/guillermo/gogo/internal/realfs"
)

// ChangeInfo contains information about a file modification
type ChangeInfo struct {
	Action     string // "create", "modify", "delete"
	FileName   string
	OldContent []byte
	NewContent []byte
	Diff       string
}

// ConflictFunc is called before applying changes
// Returns true to apply changes, false to skip
type ConflictFunc func(fs fs.FS, oldPath, newPath string, info ChangeInfo) bool

// Predefined conflict resolution strategies
var (
	// ConflictAsk prompts the user in the terminal to accept or reject each change.
	// This is the default behavior and shows diffs before asking for confirmation.
	ConflictAsk ConflictFunc = func(fs fs.FS, oldPath, newPath string, info ChangeInfo) bool {
		// Print the change information
		fmt.Printf("\n=== File: %s ===\n", info.FileName)
		fmt.Printf("Action: %s\n", info.Action)

		// Show the diff
		if info.Diff != "" {
			fmt.Println("\nChanges to be applied:")
			fmt.Println(info.Diff)
		} else if info.Action == "create" {
			fmt.Println("\nNew file will be created with content:")
			if len(info.NewContent) > 500 {
				fmt.Printf("%s\n... (truncated, %d bytes total)\n", string(info.NewContent[:500]), len(info.NewContent))
			} else {
				fmt.Println(string(info.NewContent))
			}
		}

		// Ask for confirmation
		fmt.Print("\nApply these changes? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return false
		}

		response = strings.TrimSpace(strings.ToLower(response))
		return response == "y" || response == "yes"
	}

	// ConflictAccept automatically accepts all changes without prompting.
	// Useful for testing, automation, or when you trust all modifications.
	ConflictAccept ConflictFunc = func(fs fs.FS, oldPath, newPath string, info ChangeInfo) bool {
		return true
	}

	// ConflictReject automatically rejects all changes without prompting.
	// Useful for dry-run mode or when you want to see what would be changed without applying changes.
	ConflictReject ConflictFunc = func(fs fs.FS, oldPath, newPath string, info ChangeInfo) bool {
		return false
	}
)

// Options contains options for creating a project
type Options struct {
	InitialPackageName string       // Default package name if not set
	ConflictFunc       ConflictFunc // Conflict resolution function (nil defaults to ConflictAccept)
	FS                 fs.FS        // Filesystem to use (required)
}

// StructOpts contains options for creating or modifying a struct
type StructOpts struct {
	Filename string        // File to create/modify the struct in
	Name     string        // Name of the struct
	Fields   []StructField // Fields to ensure exist (mutually exclusive with Content)
	Content  string        // Raw field content as string (mutually exclusive with Fields)

	// Optional fields for advanced usage
	DeleteFields     []StructField // Fields to remove (only used with Fields)
	PreserveExisting bool          // Preserve existing fields not mentioned (only used with Fields)
}

// MethodOpts contains options for creating or modifying a method
type MethodOpts struct {
	Filename     string      // File to create/modify the method in
	Name         string      // Name of the method
	ReceiverName string      // Receiver variable name (e.g., "u")
	ReceiverType string      // Receiver type (e.g., "User", "*User")
	Parameters   []Parameter // Method parameters (mutually exclusive with Content)
	ReturnType   string      // Return type (e.g., "error", "(string, error)")
	Body         string      // Method body content
	Content      string      // Raw method signature + body (mutually exclusive with Parameters)

	// Optional fields for advanced usage
	PreserveExisting bool // Preserve existing methods not mentioned
}

// FunctionOpts contains options for creating or modifying a function
type FunctionOpts struct {
	Filename   string      // File to create/modify the function in
	Name       string      // Name of the function
	Parameters []Parameter // Function parameters (mutually exclusive with Content)
	ReturnType string      // Return type (e.g., "error", "(string, error)")
	Body       string      // Function body content
	Content    string      // Raw function signature + body (mutually exclusive with Parameters)

	// Optional fields for advanced usage
	PreserveExisting bool // Preserve existing functions not mentioned
}

// VariableOpts contains options for creating or modifying variables
type VariableOpts struct {
	Filename  string     // File to create/modify the variables in
	Variables []Variable // Variables to ensure exist (mutually exclusive with Content)
	Content   string     // Raw variable declarations as string (mutually exclusive with Variables)

	// Optional fields for advanced usage
	DeleteVariables  []string // Variable names to remove (only used with Variables)
	PreserveExisting bool     // Preserve existing variables not mentioned (only used with Variables)
}

// ConstantOpts contains options for creating or modifying constants
type ConstantOpts struct {
	Filename  string     // File to create/modify the constants in
	Constants []Constant // Constants to ensure exist (mutually exclusive with Content)
	Content   string     // Raw constant declarations as string (mutually exclusive with Constants)

	// Optional fields for advanced usage
	DeleteConstants  []string // Constant names to remove (only used with Constants)
	PreserveExisting bool     // Preserve existing constants not mentioned (only used with Constants)
}

// TypeOpts contains options for creating or modifying type definitions
type TypeOpts struct {
	Filename string    // File to create/modify the types in
	Types    []TypeDef // Type definitions to ensure exist (mutually exclusive with Content)
	Content  string    // Raw type declarations as string (mutually exclusive with Types)

	// Optional fields for advanced usage
	DeleteTypes      []string // Type names to remove (only used with Types)
	PreserveExisting bool     // Preserve existing types not mentioned (only used with Types)
}

// structDef represents a Go struct definition (legacy - for backward compatibility)
type structDef struct {
	Name             string
	EnsureFields     []StructField // Fields to ensure exist
	DeleteFields     []StructField // Fields to remove
	PreserveExisting bool          // Preserve existing fields not mentioned
}

// StructField represents a field in a Go struct
type StructField struct {
	Name       string
	Type       string
	Annotation string // e.g., `json:"id"`
}

// Parameter represents a function/method parameter
type Parameter struct {
	Name string
	Type string
}

// Variable represents a variable declaration
type Variable struct {
	Name  string
	Type  string // Optional for inferred types
	Value string // Initial value
}

// Constant represents a constant declaration
type Constant struct {
	Name  string
	Type  string // Optional for inferred types
	Value string // Constant value
}

// TypeDef represents a type definition
type TypeDef struct {
	Name       string
	Definition string // e.g., "string", "struct { ... }", "interface { ... }"
}

// OpenFS creates a filesystem instance for the given path.
// If the directory doesn't exist, it will be created recursively.
func OpenFS(path string) (fs.FS, error) {
	return realfs.Open(path)
}

// NewFS creates a new project with a filesystem for the given path.
// If the directory doesn't exist, it will be created recursively.
// This is a convenience function that combines OpenFS and New.
func NewFS(path string, opts Options) (*Project, error) {
	filesystem, err := OpenFS(path)
	if err != nil {
		return nil, err
	}

	// Set the filesystem in the options
	opts.FS = filesystem

	return New(opts)
}

// New creates a new project
func New(opts Options) (*Project, error) {
	if opts.FS == nil {
		return nil, fmt.Errorf("filesystem is required")
	}

	// Default conflict function if not provided
	if opts.ConflictFunc == nil {
		opts.ConflictFunc = ConflictAsk
	}

	return &Project{
		opts:         opts,
		fs:           opts.FS,
		conflictFunc: opts.ConflictFunc,
	}, nil
}
