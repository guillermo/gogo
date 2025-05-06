package astm

import (
	"testing"
)

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
