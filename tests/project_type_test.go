package tests

import (
	"testing"

	"github.com/guillermo/gogo"
	"github.com/guillermo/gogo/gogotest"
)

func TestProjectType(t *testing.T) {
	t.Run("CreateSingleType", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "types",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Type(gogo.TypeOpts{
			Filename: "custom.go",
			Types: []gogo.TypeDef{
				{Name: "UserID", Definition: "string"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify the type was created
		if err := fs.Assert(`type UserID string`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("CreateMultipleTypes", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "identifiers",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Type(gogo.TypeOpts{
			Filename: "ids.go",
			Types: []gogo.TypeDef{
				{Name: "UserID", Definition: "string"},
				{Name: "ProductID", Definition: "int"},
				{Name: "OrderID", Definition: "uint64"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify all types were created
		if err := fs.Assert(`type UserID string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type ProductID int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type OrderID uint64`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FunctionType", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "handlers",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Type(gogo.TypeOpts{
			Filename: "handlers.go",
			Types: []gogo.TypeDef{
				{Name: "HandlerFunc", Definition: "func(http.ResponseWriter, *http.Request)"},
				{Name: "MiddlewareFunc", Definition: "func(HandlerFunc) HandlerFunc"},
				{Name: "ValidationFunc", Definition: "func(interface{}) error"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify function types
		if err := fs.Assert(`type HandlerFunc func(http.ResponseWriter, *http.Request)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type MiddlewareFunc func(HandlerFunc) HandlerFunc`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type ValidationFunc func(interface{}) error`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("InterfaceType", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "interfaces",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Type(gogo.TypeOpts{
			Filename: "storage.go",
			Content: `type Storage interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	Delete(key string) error
	List() ([]string, error)
}

type Cache interface {
	Storage
	Expire(key string, duration time.Duration) error
	TTL(key string) (time.Duration, error)
}`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify interface types
		if err := fs.Assert(`type Storage interface`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Get(key string) (interface{}, error)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Set(key string, value interface{}) error`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type Cache interface`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Storage`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Expire(key string, duration time.Duration) error`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("StructType", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "models",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Type(gogo.TypeOpts{
			Filename: "embedded.go",
			Content: `type BaseModel struct {
	ID        string    ` + "`json:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

type User struct {
	BaseModel
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify struct types
		if err := fs.Assert(`type BaseModel struct`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`ID        string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type User struct`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`BaseModel`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Name  string`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("SliceAndMapTypes", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "collections",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Type(gogo.TypeOpts{
			Filename: "collections.go",
			Types: []gogo.TypeDef{
				{Name: "UserList", Definition: "[]User"},
				{Name: "UserMap", Definition: "map[string]*User"},
				{Name: "StringSet", Definition: "map[string]bool"},
				{Name: "IntSlice", Definition: "[]int"},
				{Name: "NestedMap", Definition: "map[string]map[string]interface{}"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify collection types
		if err := fs.Assert(`type UserList []User`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type UserMap map[string]*User`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type StringSet map[string]bool`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type IntSlice []int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type NestedMap map[string]map[string]interface{}`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ChannelTypes", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "channels",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Type(gogo.TypeOpts{
			Filename: "channels.go",
			Types: []gogo.TypeDef{
				{Name: "JobChannel", Definition: "chan Job"},
				{Name: "ResultChannel", Definition: "chan Result"},
				{Name: "ReadOnlyChannel", Definition: "<-chan string"},
				{Name: "WriteOnlyChannel", Definition: "chan<- string"},
				{Name: "BufferedChannel", Definition: "chan int"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify channel types
		if err := fs.Assert(`type JobChannel chan Job`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type ResultChannel chan Result`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type ReadOnlyChannel <-chan string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type WriteOnlyChannel chan<- string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type BufferedChannel chan int`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("PointerTypes", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "pointers",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Type(gogo.TypeOpts{
			Filename: "pointers.go",
			Types: []gogo.TypeDef{
				{Name: "UserPtr", Definition: "*User"},
				{Name: "StringPtr", Definition: "*string"},
				{Name: "IntPtr", Definition: "*int"},
				{Name: "SlicePtr", Definition: "*[]string"},
				{Name: "MapPtr", Definition: "*map[string]interface{}"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify pointer types
		if err := fs.Assert(`type UserPtr *User`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type StringPtr *string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type IntPtr *int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type SlicePtr *[]string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type MapPtr *map[string]interface{}`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GenericTypes", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "generics",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Type(gogo.TypeOpts{
			Filename: "generics.go",
			Content: `type List[T any] struct {
	items []T
}

type Map[K comparable, V any] struct {
	data map[K]V
}

type Result[T any] struct {
	Value T
	Error error
}

type Processor[T, U any] interface {
	Process(T) (U, error)
}`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify generic types
		if err := fs.Assert(`type List[T any] struct`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`items []T`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type Map[K comparable, V any] struct`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`data map[K]V`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type Result[T any] struct`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type Processor[T, U any] interface`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Process(T) (U, error)`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("TypeBlock", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "types",
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Type(gogo.TypeOpts{
			Filename: "block.go",
			Content: `type (
	// Identifier types
	UserID    string
	ProductID int
	OrderID   uint64

	// Function types
	Handler    func(http.ResponseWriter, *http.Request)
	Middleware func(Handler) Handler
	Validator  func(interface{}) error

	// Collection types
	Users    []User
	Products map[string]Product
)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify type block
		if err := fs.Assert(`// Identifier types`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`UserID    string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`ProductID int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`// Function types`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Handler    func(http.ResponseWriter, *http.Request)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`// Collection types`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`Products map[string]Product`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MultipleTypeFiles", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "app",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create types in first file
		err = project.Type(gogo.TypeOpts{
			Filename: "user.go",
			Types: []gogo.TypeDef{
				{Name: "UserID", Definition: "string"},
				{Name: "UserRole", Definition: "string"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create types in second file
		err = project.Type(gogo.TypeOpts{
			Filename: "product.go",
			Types: []gogo.TypeDef{
				{Name: "ProductID", Definition: "int"},
				{Name: "Price", Definition: "float64"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify types in different files exist
		if err := fs.Assert(`type UserID string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type ProductID int`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("TypeWithMethods", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "custom",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create a custom type
		err = project.Type(gogo.TypeOpts{
			Filename: "status.go",
			Types: []gogo.TypeDef{
				{Name: "Status", Definition: "string"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add methods to the type
		err = project.Method(gogo.MethodOpts{
			Filename:     "status.go",
			Name:         "String",
			ReceiverType: "Status",
			ReturnType:   "string",
			Body:         `return string(s)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		err = project.Method(gogo.MethodOpts{
			Filename:     "status.go",
			Name:         "IsValid",
			ReceiverType: "Status",
			ReturnType:   "bool",
			Body:         `return s == "active" || s == "inactive"`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify type and methods
		if err := fs.Assert(`type Status string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`func (s Status) String() string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`func (s Status) IsValid() bool`); err != nil {
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
		err = project.Type(gogo.TypeOpts{
			Types: []gogo.TypeDef{
				{Name: "Test", Definition: "string"},
			},
		})
		if err == nil {
			t.Fatal("Expected error for missing filename")
		}

		// Both Types and Content provided
		err = project.Type(gogo.TypeOpts{
			Filename: "test.go",
			Types: []gogo.TypeDef{
				{Name: "Test", Definition: "string"},
			},
			Content: `type Test string`,
		})
		if err == nil {
			t.Fatal("Expected error for mutually exclusive Types and Content")
		}

		// Neither Types nor Content provided
		err = project.Type(gogo.TypeOpts{
			Filename: "test.go",
		})
		if err == nil {
			t.Fatal("Expected error for missing Types or Content")
		}
	})

	t.Run("AddToExistingFile", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "types",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create initial types
		err = project.Type(gogo.TypeOpts{
			Filename: "identifiers.go",
			Types: []gogo.TypeDef{
				{Name: "UserID", Definition: "string"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add more types to the same file
		err = project.Type(gogo.TypeOpts{
			Filename: "identifiers.go",
			Types: []gogo.TypeDef{
				{Name: "ProductID", Definition: "int"},
				{Name: "OrderID", Definition: "uint64"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify all types exist
		if err := fs.Assert(`type UserID string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type ProductID int`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`type OrderID uint64`); err != nil {
			t.Fatal(err)
		}
	})
}
