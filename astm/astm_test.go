package astm

import (
	"bytes"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name:    "valid code",
			src:     "package main\nfunc main() {}",
			wantErr: false,
		},
		{
			name:    "invalid code",
			src:     "invalid code",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse([]byte(tt.src))
			if (code == nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", code == nil, tt.wantErr)
			}
		})
	}
}

func TestPackage(t *testing.T) {
	code := Parse([]byte("package main\nfunc main() {}"))
	if code == nil {
		t.Fatal("Parse failed")
	}

	code.Package("newpkg")
	if code.file.Name.Name != "newpkg" {
		t.Errorf("Package() = %v, want %v", code.file.Name.Name, "newpkg")
	}
}

func TestRename(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		old     string
		new     string
		wantErr bool
	}{
		{
			name:    "rename function",
			src:     "package main\nfunc oldFunc() {}\nfunc main() { oldFunc() }",
			old:     "oldFunc",
			new:     "newFunc",
			wantErr: false,
		},
		{
			name:    "rename non-existent",
			src:     "package main\nfunc main() {}",
			old:     "nonexistent",
			new:     "new",
			wantErr: false,
		},
		{
			name:    "rename struct",
			src:     "package main\ntype OldStruct struct{}\nfunc main() { var s OldStruct; _ = s }",
			old:     "OldStruct",
			new:     "NewStruct",
			wantErr: false,
		},
		{
			name:    "rename variable",
			src:     "package main\nvar oldVar = 42\nfunc main() { _ = oldVar }",
			old:     "oldVar",
			new:     "newVar",
			wantErr: false,
		},
		{
			name:    "rename constant",
			src:     "package main\nconst OldConst = 42\nfunc main() { _ = OldConst }",
			old:     "OldConst",
			new:     "NewConst",
			wantErr: false,
		},
		{
			name:    "rename struct with references",
			src:     "package main\ntype OldStruct struct{}\nfunc f(s OldStruct) OldStruct { return s }",
			old:     "OldStruct",
			new:     "NewStruct",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse([]byte(tt.src))
			if code == nil {
				t.Fatal("Parse failed")
			}

			err := code.Rename(tt.old, tt.new)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rename() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify the rename was successful
			var buf bytes.Buffer
			if err := code.WriteTo(&buf); err != nil {
				t.Fatal(err)
			}
			output := buf.String()
			t.Logf("Output after rename:\n%s", output)
			if !tt.wantErr && bytes.Contains(buf.Bytes(), []byte(tt.old)) {
				t.Errorf("Rename() did not replace %v with %v", tt.old, tt.new)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		newCode string
		check   []string
		wantErr bool
	}{
		{
			name:    "set function",
			src:     "package main\nfunc main() {}",
			newCode: "package main\n\nfunc newFunc() int { return 42 }",
			check:   []string{"newFunc", "return 42"},
			wantErr: false,
		},
		{
			name:    "set variable",
			src:     "package main\nfunc main() {}",
			newCode: "package main\n\nvar newVar = 42",
			check:   []string{"newVar", "= 42"},
			wantErr: false,
		},
		{
			name:    "set constant",
			src:     "package main\nfunc main() {}",
			newCode: "package main\n\nconst NewConst = 3.14",
			check:   []string{"NewConst", "= 3.14"},
			wantErr: false,
		},
		{
			name:    "set struct",
			src:     "package main\nfunc main() {}",
			newCode: "package main\n\ntype NewStruct struct { Field int }",
			check:   []string{"NewStruct", "Field int"},
			wantErr: false,
		},
		{
			name:    "set multiple declarations",
			src:     "package main\nfunc main() {}",
			newCode: "package main\n\nconst PI = 3.14\nvar globalVar = \"value\"\ntype NewStruct struct { Field int }\nfunc newFunc() int { return 42 }",
			check:   []string{"PI", "globalVar", "NewStruct", "newFunc"},
			wantErr: false,
		},
		{
			name:    "invalid code",
			src:     "package main\nfunc main() {}",
			newCode: "invalid code",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse([]byte(tt.src))
			if code == nil {
				t.Fatal("Parse failed")
			}

			err := code.Set([]byte(tt.newCode))
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				var buf bytes.Buffer
				if err := code.WriteTo(&buf); err != nil {
					t.Fatal(err)
				}
				output := buf.String()
				t.Logf("Output after set:\n%s", output)
				// Check that the new code was added
				for _, check := range tt.check {
					if !bytes.Contains(buf.Bytes(), []byte(check)) {
						t.Errorf("Set() did not add %q", check)
					}
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		toDelete string
		wantErr  bool
	}{
		{
			name:     "delete unused function",
			src:      "package main\nfunc toDelete() {}\nfunc main() {}",
			toDelete: "toDelete",
			wantErr:  false,
		},
		{
			name:     "delete function with reference",
			src:      "package main\nfunc toDelete() {}\nfunc main() { toDelete() }",
			toDelete: "toDelete",
			wantErr:  true,
		},
		{
			name:     "delete unused struct",
			src:      "package main\ntype ToDelete struct{}\nfunc main() {}",
			toDelete: "ToDelete",
			wantErr:  false,
		},
		{
			name:     "delete struct with reference",
			src:      "package main\ntype ToDelete struct{}\nfunc main() { var s ToDelete; _ = s }",
			toDelete: "ToDelete",
			wantErr:  true,
		},
		{
			name:     "delete unused variable",
			src:      "package main\nvar toDelete = 42\nfunc main() {}",
			toDelete: "toDelete",
			wantErr:  false,
		},
		{
			name:     "delete variable with reference",
			src:      "package main\nvar toDelete = 42\nfunc main() { _ = toDelete }",
			toDelete: "toDelete",
			wantErr:  true,
		},
		{
			name:     "delete unused constant",
			src:      "package main\nconst ToDelete = 42\nfunc main() {}",
			toDelete: "ToDelete",
			wantErr:  false,
		},
		{
			name:     "delete constant with reference",
			src:      "package main\nconst ToDelete = 42\nfunc main() { _ = ToDelete }",
			toDelete: "ToDelete",
			wantErr:  true,
		},
		{
			name:     "delete non-existent",
			src:      "package main\nfunc main() {}",
			toDelete: "nonexistent",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse([]byte(tt.src))
			if code == nil {
				t.Fatal("Parse failed")
			}

			err := code.Delete(tt.toDelete)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				var buf bytes.Buffer
				if err := code.WriteTo(&buf); err != nil {
					t.Fatal(err)
				}
				output := buf.String()
				t.Logf("Output after delete:\n%s", output)
				if bytes.Contains(buf.Bytes(), []byte(tt.toDelete)) {
					t.Errorf("Delete() did not remove %v", tt.toDelete)
				}
			}
		})
	}
}

func TestFieldSet(t *testing.T) {
	code := Parse([]byte("package main\ntype MyStruct struct {}"))
	if code == nil {
		t.Fatal("Parse failed")
	}

	if err := code.FieldSet("MyStruct", "NewField", "int", "`json:\"new_field\"`"); err != nil {
		t.Errorf("FieldSet() error = %v", err)
	}

	var buf bytes.Buffer
	if err := code.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("NewField")) {
		t.Error("FieldSet() did not add the new field")
	}
}

func TestFieldRename(t *testing.T) {
	code := Parse([]byte("package main\ntype MyStruct struct { OldField int }\nfunc main() { var s MyStruct; s.OldField = 1 }"))
	if code == nil {
		t.Fatal("Parse failed")
	}

	if err := code.FieldRename("MyStruct", "OldField", "NewField"); err != nil {
		t.Errorf("FieldRename() error = %v", err)
	}

	var buf bytes.Buffer
	if err := code.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	t.Logf("Output after rename:\n%s", output)
	if bytes.Contains(buf.Bytes(), []byte("OldField")) {
		t.Error("FieldRename() did not rename the field")
	}
	if !bytes.Contains(buf.Bytes(), []byte("NewField")) {
		t.Error("FieldRename() did not add the new field name")
	}
}

func TestFieldDelete(t *testing.T) {
	code := Parse([]byte("package main\ntype MyStruct struct { ToDelete int }\nfunc main() { s := MyStruct{ToDelete: 1} }"))
	if code == nil {
		t.Fatal("Parse failed")
	}

	if err := code.FieldDelete("MyStruct", "ToDelete"); err != nil {
		t.Errorf("FieldDelete() error = %v", err)
	}

	var buf bytes.Buffer
	if err := code.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(buf.Bytes(), []byte("ToDelete")) {
		t.Error("FieldDelete() did not remove the field")
	}
}

func TestBuildTags(t *testing.T) {
	code := Parse([]byte("package main\nfunc main() {}"))
	if code == nil {
		t.Fatal("Parse failed")
	}

	tags := []string{"linux", "amd64"}
	code.BuildTags(tags)

	var buf bytes.Buffer
	if err := code.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	expected := "//go:build linux && amd64"
	if !bytes.Contains(buf.Bytes(), []byte(expected)) {
		t.Errorf("BuildTags() did not add the expected tags. Got: %s", buf.String())
	}
}

func TestWriteTo(t *testing.T) {
	code := Parse([]byte("package main\nfunc main() {}"))
	if code == nil {
		t.Fatal("Parse failed")
	}

	var buf bytes.Buffer
	if err := code.WriteTo(&buf); err != nil {
		t.Errorf("WriteTo() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("WriteTo() did not write any data")
	}
}
