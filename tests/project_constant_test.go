package tests

import (
	"testing"

	"github.com/guillermo/gogo"
	"github.com/guillermo/gogo/gogotest"
)

func TestProjectConstant(t *testing.T) {
	t.Run("CreateSingleConstant", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "config",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Constant(gogo.ConstantOpts{
			Filename: "constants.go",
			Constants: []gogo.Constant{
				{Name: "MaxRetries", Type: "int", Value: "3"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify the constant was created
		if err := fs.Assert(`const MaxRetries int = 3`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("CreateMultipleConstants", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Constant(gogo.ConstantOpts{
			Filename: "app.go",
			Constants: []gogo.Constant{
				{Name: "AppName", Type: "string", Value: `"MyApplication"`},
				{Name: "Version", Type: "string", Value: `"1.0.0"`},
				{Name: "MaxUsers", Type: "int", Value: "1000"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify all constants were created
		if err := fs.Assert(`const AppName string = "MyApplication"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const Version string = "1.0.0"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const MaxUsers int = 1000`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ConstantWithoutType", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Constant(gogo.ConstantOpts{
			Filename: "inferred.go",
			Constants: []gogo.Constant{
				{Name: "Pi", Value: "3.14159"},           // Type inferred from value
				{Name: "DefaultMessage", Value: `"Hi!"`}, // Type inferred from value
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify constants with inferred types
		if err := fs.Assert(`const Pi = 3.14159`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const DefaultMessage = "Hi!"`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ConstantWithContent", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "http",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Constant(gogo.ConstantOpts{
			Filename: "status.go",
			Content: `const (
	// HTTP Status Codes
	StatusOK           = 200
	StatusNotFound     = 404
	StatusServerError  = 500

	// Custom Status Messages
	MessageSuccess = "Operation completed successfully"
	MessageError   = "An error occurred"
)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify constants from content
		if err := fs.Assert(`// HTTP Status Codes`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`StatusOK           = 200`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`StatusNotFound     = 404`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`MessageSuccess = "Operation completed successfully"`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("IotaConstants", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "enums",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Constant(gogo.ConstantOpts{
			Filename: "status.go",
			Content: `const (
	StatusPending = iota
	StatusActive
	StatusInactive
	StatusDeleted
)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify iota constants
		if err := fs.Assert(`StatusPending = iota`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`StatusActive`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`StatusInactive`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`StatusDeleted`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("TypedIotaConstants", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "types",
		})
		if err != nil {
			t.Fatal(err)
		}

		// First create a custom type
		err = project.Type(gogo.TypeOpts{
			Filename: "priority.go",
			Types: []gogo.TypeDef{
				{Name: "Priority", Definition: "int"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Then create typed constants
		err = project.Constant(gogo.ConstantOpts{
			Filename: "priority.go",
			Content: `const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify type and constants
		if err := fs.Assert(`type Priority int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`PriorityLow Priority = iota`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`PriorityMedium`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`PriorityCritical`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ComplexConstantValues", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "math",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Constant(gogo.ConstantOpts{
			Filename: "calculations.go",
			Constants: []gogo.Constant{
				{Name: "Pi", Type: "float64", Value: "3.141592653589793"},
				{Name: "E", Type: "float64", Value: "2.718281828459045"},
				{Name: "GoldenRatio", Type: "float64", Value: "1.618033988749895"},
				{Name: "BytesPerKB", Type: "int", Value: "1024"},
				{Name: "SecondsPerDay", Type: "int", Value: "24 * 60 * 60"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify complex constant values
		if err := fs.Assert(`const Pi float64 = 3.141592653589793`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const E float64 = 2.718281828459045`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const SecondsPerDay int = 24 * 60 * 60`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("StringConstants", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "messages",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Constant(gogo.ConstantOpts{
			Filename: "strings.go",
			Constants: []gogo.Constant{
				{Name: "WelcomeMessage", Type: "string", Value: `"Welcome to our application!"`},
				{Name: "ErrorMessage", Type: "string", Value: `"Something went wrong"`},
				{Name: "APIEndpoint", Type: "string", Value: `"https://api.example.com/v1"`},
				{Name: "DateFormat", Type: "string", Value: `"2006-01-02 15:04:05"`},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify string constants
		if err := fs.Assert(`const WelcomeMessage string = "Welcome to our application!"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const APIEndpoint string = "https://api.example.com/v1"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const DateFormat string = "2006-01-02 15:04:05"`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ConstantBlock", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "config",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Constant(gogo.ConstantOpts{
			Filename: "limits.go",
			Content: `const (
	// User limits
	MaxUsernameLength = 50
	MinPasswordLength = 8
	MaxPasswordLength = 128

	// System limits
	MaxFileSize     = 10 * 1024 * 1024 // 10MB
	MaxUploadFiles  = 5
	DefaultTimeout  = 30 // seconds
)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify constant block
		if err := fs.Assert(`// User limits`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`MaxUsernameLength = 50`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`MinPasswordLength = 8`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`// System limits`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`MaxFileSize     = 10 * 1024 * 1024 // 10MB`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MultipleConstantFiles", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "app",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create constants in first file
		err = project.Constant(gogo.ConstantOpts{
			Filename: "server.go",
			Constants: []gogo.Constant{
				{Name: "DefaultPort", Type: "int", Value: "8080"},
				{Name: "DefaultHost", Type: "string", Value: `"localhost"`},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create constants in second file
		err = project.Constant(gogo.ConstantOpts{
			Filename: "database.go",
			Constants: []gogo.Constant{
				{Name: "MaxConnections", Type: "int", Value: "100"},
				{Name: "ConnectionTimeout", Type: "int", Value: "30"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify constants in different files exist
		if err := fs.Assert(`const DefaultPort int = 8080`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const MaxConnections int = 100`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("BooleanConstants", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "features",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Constant(gogo.ConstantOpts{
			Filename: "flags.go",
			Constants: []gogo.Constant{
				{Name: "EnableLogging", Type: "bool", Value: "true"},
				{Name: "EnableMetrics", Type: "bool", Value: "false"},
				{Name: "EnableCache", Type: "bool", Value: "true"},
				{Name: "DebugMode", Type: "bool", Value: "false"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify boolean constants
		if err := fs.Assert(`const EnableLogging bool = true`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const EnableMetrics bool = false`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const EnableCache bool = true`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const DebugMode bool = false`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Missing filename
		err = project.Constant(gogo.ConstantOpts{
			Constants: []gogo.Constant{
				{Name: "Test", Type: "string", Value: `"test"`},
			},
		})
		if err == nil {
			t.Fatal("Expected error for missing filename")
		}

		// Both Constants and Content provided
		err = project.Constant(gogo.ConstantOpts{
			Filename: "test.go",
			Constants: []gogo.Constant{
				{Name: "Test", Type: "string", Value: `"test"`},
			},
			Content: `const Test = "test"`,
		})
		if err == nil {
			t.Fatal("Expected error for mutually exclusive Constants and Content")
		}

		// Neither Constants nor Content provided
		err = project.Constant(gogo.ConstantOpts{
			Filename: "test.go",
		})
		if err == nil {
			t.Fatal("Expected error for missing Constants or Content")
		}
	})

	t.Run("AddToExistingFile", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "config",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create initial constants
		err = project.Constant(gogo.ConstantOpts{
			Filename: "app.go",
			Constants: []gogo.Constant{
				{Name: "AppName", Type: "string", Value: `"MyApp"`},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add more constants to the same file
		err = project.Constant(gogo.ConstantOpts{
			Filename: "app.go",
			Constants: []gogo.Constant{
				{Name: "Version", Type: "string", Value: `"2.0.0"`},
				{Name: "BuildNumber", Type: "int", Value: "42"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify all constants exist
		if err := fs.Assert(`const AppName string = "MyApp"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const Version string = "2.0.0"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const BuildNumber int = 42`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DurationConstants", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "timeouts",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Constant(gogo.ConstantOpts{
			Filename: "durations.go",
			Constants: []gogo.Constant{
				{Name: "RequestTimeout", Type: "time.Duration", Value: "30 * time.Second"},
				{Name: "ConnectionTimeout", Type: "time.Duration", Value: "5 * time.Second"},
				{Name: "ReadTimeout", Type: "time.Duration", Value: "10 * time.Second"},
				{Name: "WriteTimeout", Type: "time.Duration", Value: "10 * time.Second"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify duration constants
		if err := fs.Assert(`const RequestTimeout time.Duration = 30 * time.Second`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const ConnectionTimeout time.Duration = 5 * time.Second`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const ReadTimeout time.Duration = 10 * time.Second`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`const WriteTimeout time.Duration = 10 * time.Second`); err != nil {
			t.Fatal(err)
		}
	})
}
