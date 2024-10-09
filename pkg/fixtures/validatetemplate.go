package fixtures

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func ValidateContentAgainstFixture(generatedContent []byte, fixturePath string) error {
	// Read the fixture content
	fixtureContent, err := os.ReadFile(fixturePath)
	if err != nil {
		return fmt.Errorf("failed to read fixture: %w", err)
	}

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
