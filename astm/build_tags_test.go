package astm

import (
	"bytes"
	"testing"
)

func TestBuildTags(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		tags    []string
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "set build tags",
			src:     "package main\n\nfunc main() {}",
			tags:    []string{"linux", "amd64"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !bytes.Contains([]byte(output), []byte("//go:build linux && amd64")) {
					t.Error("BuildTags() did not add build tags")
				}
			},
		},
		{
			name:    "update existing build tags",
			src:     "//go:build linux\n\npackage main\n\nfunc main() {}",
			tags:    []string{"windows", "amd64"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !bytes.Contains([]byte(output), []byte("//go:build windows && amd64")) {
					t.Error("BuildTags() did not update build tags")
				}
			},
		},
		{
			name:    "remove all build tags",
			src:     "//go:build linux\n\npackage main\n\nfunc main() {}",
			tags:    nil,
			wantErr: false,
			check: func(t *testing.T, output string) {
				if bytes.Contains([]byte(output), []byte("//go:build")) {
					t.Error("BuildTags() did not remove build tags")
				}
			},
		},
		{
			name:    "invalid build tag",
			src:     "package main\n\nfunc main() {}",
			tags:    []string{"invalid tag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Parse([]byte(tt.src))
			if code == nil {
				t.Fatal("Parse failed")
			}

			code.BuildTags(tt.tags)
			if !tt.wantErr {
				var buf bytes.Buffer
				if err := code.WriteTo(&buf); err != nil {
					t.Fatal(err)
				}
				output := buf.String()
				t.Logf("Output after build tags:\n%s", output)
				if tt.check != nil {
					tt.check(t, output)
				}
			}
		})
	}
}
