package astm

import (
	"bytes"
	"testing"
)

func TestFieldRename(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		struct_ string
		from    string
		to      string
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "rename field",
			src:     "package main\ntype S struct { F int }\nfunc main() { var s S; s.F = 42 }",
			struct_: "S",
			from:    "F",
			to:      "G",
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !bytes.Contains([]byte(output), []byte("type S struct { G int }")) {
					t.Error("FieldRename() did not rename field in struct")
				}
				if !bytes.Contains([]byte(output), []byte("s.G = 42")) {
					t.Error("FieldRename() did not rename field in usage")
				}
			},
		},
		{
			name:    "rename field with tag",
			src:     "package main\ntype S struct { F int `json:\"f\"` }\nfunc main() { var s S; s.F = 42 }",
			struct_: "S",
			from:    "F",
			to:      "G",
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !bytes.Contains([]byte(output), []byte("type S struct { G int `json:\"f\"` }")) {
					t.Error("FieldRename() did not preserve field tag")
				}
			},
		},
		{
			name:    "rename non-existent field",
			src:     "package main\ntype S struct { F int }\nfunc main() { var s S; s.F = 42 }",
			struct_: "S",
			from:    "G",
			to:      "H",
			wantErr: true,
		},
		{
			name:    "rename field in non-existent struct",
			src:     "package main\ntype S struct { F int }\nfunc main() { var s S; s.F = 42 }",
			struct_: "T",
			from:    "F",
			to:      "G",
			wantErr: true,
		},
		{
			name:    "rename to existing field",
			src:     "package main\ntype S struct { F int; G int }\nfunc main() { var s S; s.F = 42 }",
			struct_: "S",
			from:    "F",
			to:      "G",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse([]byte(tt.src))
			if code == nil {
				t.Fatal("Parse failed")
			}

			err := code.FieldRename(tt.struct_, tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("FieldRename() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				var buf bytes.Buffer
				if err := code.WriteTo(&buf); err != nil {
					t.Fatal(err)
				}
				output := buf.String()
				t.Logf("Output after field rename:\n%s", output)
				if tt.check != nil {
					tt.check(t, output)
				}
			}
		})
	}
}
