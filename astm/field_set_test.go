package astm

import (
	"bytes"
	"testing"
)

func TestFieldSet(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		struct_  string
		field    string
		newValue string
		tags     string
		wantErr  bool
		check    func(t *testing.T, output string)
	}{
		{
			name:     "set field value",
			src:      "package main\ntype S struct { F int }\nfunc main() { var s S; s.F = 42 }",
			struct_:  "S",
			field:    "F",
			newValue: "string",
			tags:     "",
			wantErr:  false,
			check: func(t *testing.T, output string) {
				if !bytes.Contains([]byte(output), []byte("type S struct { F string }")) {
					t.Error("FieldSet() did not update field type")
				}
			},
		},
		{
			name:     "set field with tag",
			src:      "package main\ntype S struct { F int `json:\"f\"` }\nfunc main() { var s S; s.F = 42 }",
			struct_:  "S",
			field:    "F",
			newValue: "string",
			tags:     "`json:\"f\"`",
			wantErr:  false,
			check: func(t *testing.T, output string) {
				if !bytes.Contains([]byte(output), []byte("type S struct { F string `json:\"f\"` }")) {
					t.Error("FieldSet() did not preserve field tag")
				}
			},
		},
		{
			name:     "set non-existent field",
			src:      "package main\ntype S struct { F int }\nfunc main() { var s S; s.F = 42 }",
			struct_:  "S",
			field:    "G",
			newValue: "string",
			tags:     "",
			wantErr:  true,
		},
		{
			name:     "set field in non-existent struct",
			src:      "package main\ntype S struct { F int }\nfunc main() { var s S; s.F = 42 }",
			struct_:  "T",
			field:    "F",
			newValue: "string",
			tags:     "",
			wantErr:  true,
		},
		{
			name:     "set field with invalid value",
			src:      "package main\ntype S struct { F int }\nfunc main() { var s S; s.F = 42 }",
			struct_:  "S",
			field:    "F",
			newValue: "invalid syntax",
			tags:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse([]byte(tt.src))
			if code == nil {
				t.Fatal("Parse failed")
			}

			err := code.FieldSet(tt.struct_, tt.field, tt.newValue, tt.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("FieldSet() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				var buf bytes.Buffer
				if err := code.WriteTo(&buf); err != nil {
					t.Fatal(err)
				}
				output := buf.String()
				t.Logf("Output after field set:\n%s", output)
				if tt.check != nil {
					tt.check(t, output)
				}
			}
		})
	}
}
