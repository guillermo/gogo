package astm

import (
	"bytes"
	"testing"
)

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
