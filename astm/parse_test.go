package astm

import (
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
