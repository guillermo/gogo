package astm

import (
	"bytes"
	"testing"
)

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
