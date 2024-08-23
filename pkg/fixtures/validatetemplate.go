package fixtures

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

//go:embed pipelines/*
var pipelines embed.FS

func ValidateContentAgainstFixture(generatedContent []byte, fixturePath string) error {
	fullFixturePath := fmt.Sprintf("%s", fixturePath)

	// Read the fixture content
	fixtureContent, err := os.ReadFile(fullFixturePath)
	if err != nil {
		return fmt.Errorf("failed to read fixture: %w", err)
	}

	printDifferences(string(fixtureContent), string(generatedContent))

	if normalizeWhitespace(fixtureContent) != normalizeWhitespace(generatedContent) {
		return errors.New("generated content does not match fixture")
	}

	return nil
}

func normalizeWhitespace(content []byte) string {
	s := string(content)
	re := regexp.MustCompile(`\r?\n`)
	s = re.ReplaceAllString(s, "\n")
	re = regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

func printDifferences(s1, s2 string) {
	s1 = strings.ReplaceAll(s1, "\r", "")
	s2 = strings.ReplaceAll(s2, "\r", "")

	lines1 := strings.Split(s1, "\n")
	lines2 := strings.Split(s2, "\n")

	maxLines := len(lines1)
	if len(lines2) > maxLines {
		maxLines = len(lines2)
	}

	for i := 0; i < maxLines; i++ {
		var line1, line2 string
		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if line1 != line2 {
			fmt.Printf("Difference at line %d:\n", i+1)
			fmt.Printf("- [%s]\n", line1)
			fmt.Printf("+ [%s]\n", line2)

			// Print raw byte values for deeper inspection
			fmt.Printf("Raw bytes for line %d:\n", i+1)
			fmt.Printf("- %v\n", []byte(line1))
			fmt.Printf("+ %v\n", []byte(line2))
		}
	}
}
