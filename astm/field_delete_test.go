package astm

import (
	"bytes"
	"testing"
)

func TestFieldDelete(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		struct_ string
		field   string
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "delete unused field",
			src:     "package main\ntype S struct { F int; G int }\nfunc main() { var s S; s.G = 42 }",
			struct_: "S",
			field:   "F",
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !bytes.Contains([]byte(output), []byte("type S struct { G int }")) {
					t.Error("FieldDelete() did not remove field from struct")
				}
			},
		},
		{
			name:    "delete field with reference",
			src:     "package main\ntype S struct { F int; G int }\nfunc main() { var s S; s.F = 42 }",
			struct_: "S",
			field:   "F",
			wantErr: true,
		},
		{
			name:    "delete non-existent field",
			src:     "package main\ntype S struct { F int }\nfunc main() { var s S; s.F = 42 }",
			struct_: "S",
			field:   "G",
			wantErr: true,
		},
		{
			name:    "delete field in non-existent struct",
			src:     "package main\ntype S struct { F int }\nfunc main() { var s S; s.F = 42 }",
			struct_: "T",
			field:   "F",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse([]byte(tt.src))
			if code == nil {
				t.Fatal("Parse failed")
			}

			err := code.FieldDelete(tt.struct_, tt.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("FieldDelete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				var buf bytes.Buffer
				if err := code.WriteTo(&buf); err != nil {
					t.Fatal(err)
				}
				output := buf.String()
				t.Logf("Output after field delete:\n%s", output)
				if tt.check != nil {
					tt.check(t, output)
				}
			}
		})
	}
}
