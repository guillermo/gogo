package tests

import (
	"testing"

	"github.com/guillermo/gogo"
	"github.com/guillermo/gogo/gogotest"
)

func TestProjectMethod(t *testing.T) {
	t.Run("CreateNewMethod", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "models",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create a struct first
		err = project.Struct(gogo.StructOpts{
			Filename: "user.go",
			Name:     "User",
			Fields: []gogo.StructField{
				{Name: "ID", Type: "string"},
				{Name: "Name", Type: "string"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add a method
		err = project.Method(gogo.MethodOpts{
			Filename:     "user.go",
			Name:         "GetID",
			ReceiverType: "*User",
			ReturnType:   "string",
			Body:         `return u.ID`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify the method was created
		if err := fs.Assert(`func (u *User) GetID() string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`return u.ID`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MethodWithParameters", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create struct
		err = project.Struct(gogo.StructOpts{
			Filename: "model.go",
			Name:     "Product",
			Fields: []gogo.StructField{
				{Name: "Price", Type: "float64"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add method with parameters
		err = project.Method(gogo.MethodOpts{
			Filename:     "model.go",
			Name:         "ApplyDiscount",
			ReceiverType: "*Product",
			Parameters: []gogo.Parameter{
				{Name: "discount", Type: "float64"},
			},
			ReturnType: "float64",
			Body:       `return p.Price * (1 - discount)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func (p *Product) ApplyDiscount(discount float64) float64`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`return p.Price * (1 - discount)`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MethodWithContent", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "models",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create struct
		err = project.Struct(gogo.StructOpts{
			Filename: "order.go",
			Name:     "Order",
			Fields: []gogo.StructField{
				{Name: "Items", Type: "[]string"},
				{Name: "Total", Type: "float64"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add method using Content string
		err = project.Method(gogo.MethodOpts{
			Filename:     "order.go",
			Name:         "CalculateTotal",
			ReceiverName: "o",
			ReceiverType: "*Order",
			Content: `() float64 {
	var total float64
	for _, item := range o.Items {
		// Calculate item price
		total += 10.0
	}
	return total
}`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func (o *Order) CalculateTotal() float64`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`for _, item := range o.Items`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("CustomReceiverName", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create struct
		err = project.Struct(gogo.StructOpts{
			Filename: "entity.go",
			Name:     "Entity",
			Fields: []gogo.StructField{
				{Name: "Name", Type: "string"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add method with custom receiver name
		err = project.Method(gogo.MethodOpts{
			Filename:     "entity.go",
			Name:         "String",
			ReceiverName: "ent",
			ReceiverType: "Entity",
			ReturnType:   "string",
			Body:         `return ent.Name`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify custom receiver name is used
		if err := fs.Assert(`func (ent Entity) String() string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`return ent.Name`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("PointerReceiverAutoName", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create struct
		err = project.Struct(gogo.StructOpts{
			Filename: "account.go",
			Name:     "Account",
			Fields: []gogo.StructField{
				{Name: "Balance", Type: "float64"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add method without specifying receiver name (should auto-generate 'a')
		err = project.Method(gogo.MethodOpts{
			Filename:     "account.go",
			Name:         "GetBalance",
			ReceiverType: "*Account",
			ReturnType:   "float64",
			Body:         `return a.Balance`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify auto-generated receiver name
		if err := fs.Assert(`func (a *Account) GetBalance() float64`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`return a.Balance`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MultipleMethodsOnSameStruct", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "models",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create struct
		err = project.Struct(gogo.StructOpts{
			Filename: "book.go",
			Name:     "Book",
			Fields: []gogo.StructField{
				{Name: "Title", Type: "string"},
				{Name: "Author", Type: "string"},
				{Name: "Pages", Type: "int"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add first method
		err = project.Method(gogo.MethodOpts{
			Filename:     "book.go",
			Name:         "GetTitle",
			ReceiverType: "*Book",
			ReturnType:   "string",
			Body:         `return b.Title`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add second method
		err = project.Method(gogo.MethodOpts{
			Filename:     "book.go",
			Name:         "GetAuthor",
			ReceiverType: "*Book",
			ReturnType:   "string",
			Body:         `return b.Author`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add third method with parameters
		err = project.Method(gogo.MethodOpts{
			Filename:     "book.go",
			Name:         "SetPages",
			ReceiverType: "*Book",
			Parameters:   []gogo.Parameter{{Name: "pages", Type: "int"}},
			Body:         `b.Pages = pages`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify all methods exist
		if err := fs.Assert(`func (b *Book) GetTitle() string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`func (b *Book) GetAuthor() string`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`func (b *Book) SetPages(pages int)`); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MethodWithMultipleReturns", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "main",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create struct
		err = project.Struct(gogo.StructOpts{
			Filename: "cache.go",
			Name:     "Cache",
			Fields: []gogo.StructField{
				{Name: "data", Type: "map[string]interface{}"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add method with multiple return values
		err = project.Method(gogo.MethodOpts{
			Filename:     "cache.go",
			Name:         "Get",
			ReceiverType: "*Cache",
			Parameters:   []gogo.Parameter{{Name: "key", Type: "string"}},
			ReturnType:   "(interface{}, bool)",
			Body: `val, exists := c.data[key]
	return val, exists`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`func (c *Cache) Get(key string) (interface{}, bool)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`val, exists := c.data[key]`); err != nil {
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
		err = project.Method(gogo.MethodOpts{
			Name:         "Test",
			ReceiverType: "*Test",
			Body:         `return nil`,
		})
		if err == nil {
			t.Fatal("Expected error for missing filename")
		}

		// Missing method name
		err = project.Method(gogo.MethodOpts{
			Filename:     "test.go",
			ReceiverType: "*Test",
			Body:         `return nil`,
		})
		if err == nil {
			t.Fatal("Expected error for missing method name")
		}

		// Missing receiver type
		err = project.Method(gogo.MethodOpts{
			Filename: "test.go",
			Name:     "Test",
			Body:     `return nil`,
		})
		if err == nil {
			t.Fatal("Expected error for missing receiver type")
		}

		// Both structured params and Content provided
		err = project.Method(gogo.MethodOpts{
			Filename:     "test.go",
			Name:         "Test",
			ReceiverType: "*Test",
			Body:         `return nil`,
			Content:      `() { return nil }`,
		})
		if err == nil {
			t.Fatal("Expected error for mutually exclusive Body and Content")
		}

		// Neither structured params nor Content provided
		err = project.Method(gogo.MethodOpts{
			Filename:     "test.go",
			Name:         "Test",
			ReceiverType: "*Test",
		})
		if err == nil {
			t.Fatal("Expected error for missing Body or Content")
		}
	})

	t.Run("MethodOnInterfaceReceiver", func(t *testing.T) {
		fs := gogotest.New("")
		project, err := gogo.New(gogo.Options{
			FS:                 fs,
			ConflictFunc:       gogo.ConflictAccept,
			InitialPackageName: "handlers",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Create a type alias
		err = project.Type(gogo.TypeOpts{
			Filename: "handler.go",
			Types: []gogo.TypeDef{
				{Name: "HandlerFunc", Definition: "func(w http.ResponseWriter, r *http.Request)"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Add method on the type alias
		err = project.Method(gogo.MethodOpts{
			Filename:     "handler.go",
			Name:         "ServeHTTP",
			ReceiverName: "f",
			ReceiverType: "HandlerFunc",
			Parameters: []gogo.Parameter{
				{Name: "w", Type: "http.ResponseWriter"},
				{Name: "r", Type: "*http.Request"},
			},
			Body: `f(w, r)`,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Verify
		if err := fs.Assert(`type HandlerFunc func(w http.ResponseWriter, r *http.Request)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request)`); err != nil {
			t.Fatal(err)
		}
		if err := fs.Assert(`f(w, r)`); err != nil {
			t.Fatal(err)
		}
	})
}
