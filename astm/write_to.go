package astm

import (
	"go/printer"
	"io"
)

// WriteTo writes the modified code to the given writer
func (c *Code) WriteTo(w io.Writer) error {
	return printer.Fprint(w, c.fset, c.file)
}
