package gogo

import (
	"bytes"
	"fmt"
	"strings"
)

// generateDiff creates a simple diff between old and new content
func generateDiff(oldContent, newContent []byte, filename string) string {
	if len(oldContent) == 0 {
		// New file
		return fmt.Sprintf("--- /dev/null\n+++ %s\n%s", filename, prefixLines(string(newContent), "+"))
	}

	oldLines := strings.Split(string(oldContent), "\n")
	newLines := strings.Split(string(newContent), "\n")

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("--- %s\n+++ %s\n", filename, filename))

	// Simple line-by-line diff
	// This is a basic implementation - could use a proper diff algorithm
	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	inDiff := false
	contextLines := 3
	var diffBlock bytes.Buffer
	lineNum := 0

	for i := 0; i < maxLines; i++ {
		var oldLine, newLine string
		hasOld := i < len(oldLines)
		hasNew := i < len(newLines)

		if hasOld {
			oldLine = oldLines[i]
		}
		if hasNew {
			newLine = newLines[i]
		}

		if hasOld && hasNew && oldLine == newLine {
			// Lines are the same
			if inDiff {
				diffBlock.WriteString(fmt.Sprintf(" %s\n", oldLine))
			}
			lineNum++
		} else {
			// Lines differ
			if !inDiff {
				// Start new diff block
				inDiff = true
				diffBlock.Reset()

				// Add context before
				for j := i - contextLines; j < i; j++ {
					if j >= 0 && j < len(oldLines) {
						diffBlock.WriteString(fmt.Sprintf(" %s\n", oldLines[j]))
					}
				}

				buf.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", i+1, contextLines, i+1, contextLines))
			}

			if hasOld && !hasNew {
				diffBlock.WriteString(fmt.Sprintf("-%s\n", oldLine))
			} else if !hasOld && hasNew {
				diffBlock.WriteString(fmt.Sprintf("+%s\n", newLine))
			} else {
				diffBlock.WriteString(fmt.Sprintf("-%s\n", oldLine))
				diffBlock.WriteString(fmt.Sprintf("+%s\n", newLine))
			}
		}

		// Check if we should end the diff block
		if inDiff && (i == maxLines-1 || (hasOld && hasNew && oldLine == newLine && lineNum > contextLines)) {
			// Add context after
			contextAfter := 0
			for j := i + 1; j <= i+contextLines && j < maxLines; j++ {
				if j < len(oldLines) && j < len(newLines) && oldLines[j] == newLines[j] {
					diffBlock.WriteString(fmt.Sprintf(" %s\n", oldLines[j]))
					contextAfter++
				} else {
					break
				}
			}

			buf.Write(diffBlock.Bytes())
			inDiff = false
		}
	}

	return buf.String()
}

// prefixLines adds a prefix to each line of the content
func prefixLines(content, prefix string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if line != "" || i < len(lines)-1 {
			lines[i] = prefix + " " + line
		}
	}
	return strings.Join(lines, "\n")
}

// simpleDiff provides a simpler diff format for testing
func simpleDiff(oldContent, newContent []byte) []string {
	var changes []string

	oldLines := strings.Split(string(oldContent), "\n")
	newLines := strings.Split(string(newContent), "\n")

	// Track added and removed lines
	oldMap := make(map[string]int)
	for i, line := range oldLines {
		oldMap[line] = i + 1
	}

	newMap := make(map[string]int)
	for i, line := range newLines {
		newMap[line] = i + 1
	}

	// Find removed lines
	for _, line := range oldLines {
		if _, exists := newMap[line]; !exists && line != "" {
			changes = append(changes, fmt.Sprintf("- %s", line))
		}
	}

	// Find added lines
	for _, line := range newLines {
		if _, exists := oldMap[line]; !exists && line != "" {
			changes = append(changes, fmt.Sprintf("+ %s", line))
		}
	}

	return changes
}
