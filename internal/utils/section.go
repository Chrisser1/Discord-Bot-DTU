package utils

import (
	"fmt"
	"strings"
)

type Section interface {
	SetInLine(IsInLine bool)
	GetSectionName() string
	GetSectionValue() string
	GetSectionInline() bool
}

// chunkString: Utility to split a string into slices of maxLen
func ChunkString(s string, maxLen int) []string {
	var chunks []string
	for len(s) > maxLen {
		chunks = append(chunks, s[:maxLen])
		s = s[maxLen:]
	}
	if len(s) > 0 {
		chunks = append(chunks, s)
	}
	return chunks
}

// WriteLine returns a formatted line for Discord if the value is not empty.
// For example: > **Label**: `value`   or > **Label**: [Link](url)
func WriteLine(label, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}

	// Remove newlines from the value
	value = strings.ReplaceAll(value, "\n", " ")
	return fmt.Sprintf("> **%s**: %s\n", label, value)
}

// WriteTitle returns a formatted title line.
func WriteTitle(title string) string {
	if strings.TrimSpace(title) == "" {
		return ""
	}
	return fmt.Sprintf("**%s**\n", title)
}
