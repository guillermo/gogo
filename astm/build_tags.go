package astm

import (
	"go/ast"
	"strings"
)

// BuildTags sets the build tags for the file
func (c *Code) BuildTags(tags []string) {
	if len(tags) == 0 {
		return
	}

	// Create a new comment group for build tags
	comment := "//go:build " + strings.Join(tags, " && ")
	c.file.Comments = append([]*ast.CommentGroup{
		{
			List: []*ast.Comment{
				{
					Text: comment,
				},
			},
		},
	}, c.file.Comments...)
}
