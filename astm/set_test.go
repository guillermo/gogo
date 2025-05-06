package astm

import (
	"bytes"
	"testing"
)

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
