package preprocessing

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// Returns the content of a manifest file as bytes
func getManifestAsBytes(t *testing.T, filePath string) []byte {
	yamlFileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read YAML file: %s", err)
	}

	return yamlFileContent
}

// Normalize returns, newlines, extra characters with strings for easy .yaml byte comparison
func normalizeNewlines(data []byte) []byte {
	str := string(data)

	// Replace various newline characters with a single newline
	str = strings.ReplaceAll(str, "\r\n", "\n")
	str = strings.ReplaceAll(str, "\r", "\n")

	// Replace YAML block scalars' indicators and multiple spaces
	str = regexp.MustCompile(`(\s*\|\s*)`).ReplaceAllString(str, " ")
	str = strings.Join(strings.Fields(str), " ")

	// Normalize empty mappings and fields
	str = regexp.MustCompile(`\{\s*\}`).ReplaceAllString(str, "{}")
	str = regexp.MustCompile(`\s*:\s*`).ReplaceAllString(str, ": ")

	return []byte(str)
}
