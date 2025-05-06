package astm

// Package sets the package name
func (c *Code) Package(name string) {
	c.file.Name.Name = name
}
