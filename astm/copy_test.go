package astm

import (
	"bytes"
	"testing"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		from    string
		to      string
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "copy function",
			src:     "package main\nfunc original() {}\nfunc main() {}",
			from:    "original",
			to:      "copy",
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !bytes.Contains([]byte(output), []byte("func copy() {}")) {
					t.Error("Copy() did not copy function")
				}
			},
		},
		{
			name:    "copy struct type",
			src:     "package main\ntype Original struct{}\nfunc main() {}",
			from:    "Original",
			to:      "Copy",
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !bytes.Contains([]byte(output), []byte("type Copy struct{}")) {
					t.Error("Copy() did not copy struct type")
				}
			},
		},
		{
			name:    "copy variable",
			src:     "package main\nvar original = 42\nfunc main() {}",
			from:    "original",
			to:      "copy",
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !bytes.Contains([]byte(output), []byte("var copy = 42")) {
					t.Error("Copy() did not copy variable")
				}
			},
		},
		{
			name:    "copy struct field",
			src:     "package main\ntype S struct { F int }\nfunc main() { var s S; s.F = 42 }",
			from:    "F",
			to:      "G",
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !bytes.Contains([]byte(output), []byte("type S struct { F int; G int }")) {
					t.Error("Copy() did not copy struct field")
				}
			},
		},
		{
			name:    "copy non-existent",
			src:     "package main\nfunc main() {}",
			from:    "nonexistent",
			to:      "copy",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse([]byte(tt.src))
			if code == nil {
				t.Fatal("Parse failed")
			}

			err := code.Copy(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("Copy() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				var buf bytes.Buffer
				if err := code.WriteTo(&buf); err != nil {
					t.Fatal(err)
				}
				output := buf.String()
				t.Logf("Output after copy:\n%s", output)
				if tt.check != nil {
					tt.check(t, output)
				}
			}
		})
	}
}
