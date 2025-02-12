package utils

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
