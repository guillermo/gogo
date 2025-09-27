package tests

import (
	"testing"

	"github.com/guillermo/gogo"
	"github.com/guillermo/gogo/gogotest"
)

func TestProjectVariable(t *testing.T) {
	t.Run("CreateSingleVariable", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "config",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Variable(gogo.VariableOpts{
			Filename: "config.go",
			Variables: []gogo.Variable{
				{Name: "DefaultPort", Type: "int", Value: "8080"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify the variable was created
		if err := fs.Assert(`var DefaultPort int = 8080`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("CreateMultipleVariables", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Variable(gogo.VariableOpts{
			Filename: "globals.go",
			Variables: []gogo.Variable{
				{Name: "AppName", Type: "string", Value: `"MyApp"`},
				{Name: "Version", Type: "string", Value: `"1.0.0"`},
				{Name: "Debug", Type: "bool", Value: "true"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify all variables were created
		if err := fs.Assert(`var AppName string = "MyApp"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var Version string = "1.0.0"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var Debug bool = true`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("VariableWithoutType", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Variable(gogo.VariableOpts{
			Filename: "inferred.go",
			Variables: []gogo.Variable{
				{Name: "Message", Value: `"Hello, World!"`}, // Type inferred from value
				{Name: "Count", Value: "42"},                // Type inferred from value
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify variables with inferred types
		if err := fs.Assert(`var Message = "Hello, World!"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var Count = 42`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("VariableWithoutValue", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Variable(gogo.VariableOpts{
			Filename: "declarations.go",
			Variables: []gogo.Variable{
				{Name: "Logger", Type: "*log.Logger"}, // No initial value
				{Name: "DB", Type: "*sql.DB"},         // No initial value
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify variables without initial values
		if err := fs.Assert(`var Logger *log.Logger`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var DB *sql.DB`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("VariableWithContent", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "config",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Variable(gogo.VariableOpts{
			Filename: "settings.go",
			Content: `var (
	ServerHost = "localhost"
	ServerPort = 8080
	MaxConnections = 100
	EnableLogging = true
)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify variables from content
		if err := fs.Assert(`ServerHost = "localhost"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`ServerPort = 8080`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`MaxConnections = 100`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`EnableLogging = true`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ComplexVariableTypes", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "data",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Variable(gogo.VariableOpts{
			Filename: "complex.go",
			Variables: []gogo.Variable{
				{Name: "UserMap", Type: "map[string]*User", Value: "make(map[string]*User)"},
				{Name: "IdList", Type: "[]int", Value: "[]int{1, 2, 3, 4, 5}"},
				{Name: "Channel", Type: "chan string", Value: "make(chan string, 10)"},
				{Name: "Callback", Type: "func(string) error", Value: "nil"},
				{Name: "Interface", Type: "interface{}", Value: "nil"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify complex types
		if err := fs.Assert(`var UserMap map[string]*User = make(map[string]*User)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var IdList []int = []int{1, 2, 3, 4, 5}`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var Channel chan string = make(chan string, 10)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var Callback func(string) error = nil`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var Interface interface{} = nil`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GlobalVariables", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "app",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Variable(gogo.VariableOpts{
			Filename: "globals.go",
			Variables: []gogo.Variable{
				{Name: "StartTime", Type: "time.Time", Value: "time.Now()"},
				{Name: "ConfigPath", Type: "string", Value: `"/etc/app/config.json"`},
				{Name: "WorkerCount", Type: "int", Value: "runtime.NumCPU()"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify global variables
		if err := fs.Assert(`var StartTime time.Time = time.Now()`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var ConfigPath string = "/etc/app/config.json"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var WorkerCount int = runtime.NumCPU()`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("VariableBlock", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "constants",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Variable(gogo.VariableOpts{
			Filename: "block.go",
			Content: `var (
	// Database configuration
	DBHost     = "localhost"
	DBPort     = 5432
	DBName     = "myapp"
	DBUser     = "admin"
	DBPassword = "secret"

	// Redis configuration
	RedisHost = "localhost"
	RedisPort = 6379
	RedisDB   = 0
)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify variable block
		if err := fs.Assert(`// Database configuration`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`DBHost     = "localhost"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`DBPort     = 5432`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`// Redis configuration`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`RedisHost = "localhost"`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MultipleVariableFiles", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "config",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create variables in first file
		err = project.Variable(gogo.VariableOpts{
			Filename: "server.go",
			Variables: []gogo.Variable{
				{Name: "ServerPort", Type: "int", Value: "8080"},
				{Name: "ServerHost", Type: "string", Value: `"0.0.0.0"`},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create variables in second file
		err = project.Variable(gogo.VariableOpts{
			Filename: "database.go",
			Variables: []gogo.Variable{
				{Name: "DatabaseURL", Type: "string", Value: `"postgresql://localhost/mydb"`},
				{Name: "MaxConnections", Type: "int", Value: "10"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify variables in different files exist
		if err := fs.Assert(`var ServerPort int = 8080`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var DatabaseURL string = "postgresql://localhost/mydb"`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("StructVariables", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "models",
		})
		if err != nil {
			t.Fatal(err)
		}

		// First create the struct
		err = project.Struct(gogo.StructOpts{
			Filename: "config.go",
			Name:     "Config",
			Fields: []gogo.StructField{
				{Name: "Port", Type: "int"},
				{Name: "Host", Type: "string"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Then create variables using the struct
		err = project.Variable(gogo.VariableOpts{
			Filename: "config.go",
			Variables: []gogo.Variable{
				{Name: "DefaultConfig", Type: "Config", Value: `Config{Port: 8080, Host: "localhost"}`},
				{Name: "ProductionConfig", Type: "*Config", Value: `&Config{Port: 80, Host: "0.0.0.0"}`},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify struct and variables
		if err := fs.Assert(`type Config struct`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var DefaultConfig Config = Config{Port: 8080, Host: "localhost"}`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var ProductionConfig *Config = &Config{Port: 80, Host: "0.0.0.0"}`); err != nil {
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
		err = project.Variable(gogo.VariableOpts{
			Variables: []gogo.Variable{
				{Name: "Test", Type: "string", Value: `"test"`},
			},
		})
		if err == nil {
			t.Fatal("Expected error for missing filename")
		}

		// Both Variables and Content provided
		err = project.Variable(gogo.VariableOpts{
			Filename: "test.go",
			Variables: []gogo.Variable{
				{Name: "Test", Type: "string", Value: `"test"`},
			},
			Content: `var Test = "test"`,
		})
		if err == nil {
			t.Fatal("Expected error for mutually exclusive Variables and Content")
		}

		// Neither Variables nor Content provided
		err = project.Variable(gogo.VariableOpts{
			Filename: "test.go",
		})
		if err == nil {
			t.Fatal("Expected error for missing Variables or Content")
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

		// Create initial variables
		err = project.Variable(gogo.VariableOpts{
			Filename: "app.go",
			Variables: []gogo.Variable{
				{Name: "AppName", Type: "string", Value: `"MyApp"`},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add more variables to the same file
		err = project.Variable(gogo.VariableOpts{
			Filename: "app.go",
			Variables: []gogo.Variable{
				{Name: "Version", Type: "string", Value: `"2.0.0"`},
				{Name: "BuildDate", Type: "string", Value: `"2024-01-01"`},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify all variables exist
		if err := fs.Assert(`var AppName string = "MyApp"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var Version string = "2.0.0"`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`var BuildDate string = "2024-01-01"`); err != nil {
			t.Fatal(err)
		}
	})
}
