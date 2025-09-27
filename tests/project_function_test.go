package tests

import (
	"testing"

	"github.com/guillermo/gogo"
	"github.com/guillermo/gogo/gogotest"
)

func TestProjectFunction(t *testing.T) {
	t.Run("CreateNewFunction", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "utils",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Function(gogo.FunctionOpts{
			Filename:   "helpers.go",
			Name:       "Add",
			Parameters: []gogo.Parameter{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
			ReturnType: "int",
			Body:       `return a + b`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify the function was created
		if err := fs.Assert(`func Add(a int, b int) int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`return a + b`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FunctionWithNoParameters", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Function(gogo.FunctionOpts{
			Filename:   "main.go",
			Name:       "GetVersion",
			ReturnType: "string",
			Body:       `return "1.0.0"`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func GetVersion() string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`return "1.0.0"`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FunctionWithNoReturnType", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Function(gogo.FunctionOpts{
			Filename:   "logger.go",
			Name:       "LogMessage",
			Parameters: []gogo.Parameter{{Name: "message", Type: "string"}},
			Body:       `fmt.Println(message)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func LogMessage(message string)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`fmt.Println(message)`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FunctionWithContent", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "handlers",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Function(gogo.FunctionOpts{
			Filename: "middleware.go",
			Name:     "CorsMiddleware",
			Content: `(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		next.ServeHTTP(w, r)
	})
}`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func CorsMiddleware(next http.Handler) http.Handler`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`w.Header().Set("Access-Control-Allow-Origin", "*")`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`next.ServeHTTP(w, r)`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FunctionWithMultipleReturns", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "db",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Function(gogo.FunctionOpts{
			Filename:   "connection.go",
			Name:       "Connect",
			Parameters: []gogo.Parameter{{Name: "dsn", Type: "string"}},
			ReturnType: "(*sql.DB, error)",
			Body: `db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func Connect(dsn string) (*sql.DB, error)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`db, err := sql.Open("postgres", dsn)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`return db, nil`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FunctionWithComplexParameters", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "api",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Function(gogo.FunctionOpts{
			Filename: "handlers.go",
			Name:     "HandleRequest",
			Parameters: []gogo.Parameter{
				{Name: "w", Type: "http.ResponseWriter"},
				{Name: "r", Type: "*http.Request"},
				{Name: "config", Type: "map[string]interface{}"},
				{Name: "handlers", Type: "[]func(http.ResponseWriter, *http.Request)"},
			},
			ReturnType: "error",
			Body: `for _, handler := range handlers {
		handler(w, r)
	}
	return nil`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func HandleRequest(w http.ResponseWriter, r *http.Request, config map[string]interface{}, handlers []func(http.ResponseWriter, *http.Request)) error`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`for _, handler := range handlers`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MultipleFunctionsInSameFile", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "math",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add first function
		err = project.Function(gogo.FunctionOpts{
			Filename:   "operations.go",
			Name:       "Add",
			Parameters: []gogo.Parameter{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
			ReturnType: "int",
			Body:       `return a + b`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add second function
		err = project.Function(gogo.FunctionOpts{
			Filename:   "operations.go",
			Name:       "Subtract",
			Parameters: []gogo.Parameter{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
			ReturnType: "int",
			Body:       `return a - b`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add third function
		err = project.Function(gogo.FunctionOpts{
			Filename:   "operations.go",
			Name:       "Multiply",
			Parameters: []gogo.Parameter{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
			ReturnType: "int",
			Body:       `return a * b`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify all functions exist
		if err := fs.Assert(`func Add(a int, b int) int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`func Subtract(a int, b int) int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`func Multiply(a int, b int) int`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FunctionWithVariadicParameters", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "utils",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Function(gogo.FunctionOpts{
			Filename: "variadic.go",
			Name:     "Sum",
			Parameters: []gogo.Parameter{
				{Name: "numbers", Type: "...int"},
			},
			ReturnType: "int",
			Body: `var total int
	for _, num := range numbers {
		total += num
	}
	return total`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func Sum(numbers ...int) int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`for _, num := range numbers`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FunctionWithChannelParameters", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "worker",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Function(gogo.FunctionOpts{
			Filename: "worker.go",
			Name:     "Worker",
			Parameters: []gogo.Parameter{
				{Name: "jobs", Type: "<-chan Job"},
				{Name: "results", Type: "chan<- Result"},
			},
			Body: `for job := range jobs {
		result := processJob(job)
		results <- result
	}`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func Worker(jobs <-chan Job, results chan<- Result)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`for job := range jobs`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FunctionWithInterfaceParameters", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "service",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Function(gogo.FunctionOpts{
			Filename: "service.go",
			Name:     "ProcessData",
			Parameters: []gogo.Parameter{
				{Name: "processor", Type: "DataProcessor"},
				{Name: "data", Type: "interface{}"},
			},
			ReturnType: "(interface{}, error)",
			Body: `result, err := processor.Process(data)
	if err != nil {
		return nil, err
	}
	return result, nil`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func ProcessData(processor DataProcessor, data interface{}) (interface{}, error)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`result, err := processor.Process(data)`); err != nil {
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
		err = project.Function(gogo.FunctionOpts{
			Name:       "Test",
			ReturnType: "string",
			Body:       `return "test"`,
		})
		if err == nil {
			t.Fatal("Expected error for missing filename")
		}

		// Missing function name
		err = project.Function(gogo.FunctionOpts{
			Filename:   "test.go",
			ReturnType: "string",
			Body:       `return "test"`,
		})
		if err == nil {
			t.Fatal("Expected error for missing function name")
		}

		// Both structured params and Content provided
		err = project.Function(gogo.FunctionOpts{
			Filename:   "test.go",
			Name:       "Test",
			ReturnType: "string",
			Body:       `return "test"`,
			Content:    `() string { return "test" }`,
		})
		if err == nil {
			t.Fatal("Expected error for mutually exclusive Body and Content")
		}

		// Neither structured params nor Content provided
		err = project.Function(gogo.FunctionOpts{
			Filename: "test.go",
			Name:     "Test",
		})
		if err == nil {
			t.Fatal("Expected error for missing Body or Content")
		}
	})

	t.Run("MainFunction", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Function(gogo.FunctionOpts{
			Filename: "main.go",
			Name:     "main",
			Body: `fmt.Println("Hello, World!")
	os.Exit(0)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func main()`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`fmt.Println("Hello, World!")`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`os.Exit(0)`); err != nil {
			t.Fatal(err)
		}
	})
}
