package fixtures

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-cmp/cmp"
)

func ValidateContentAgainstFixture(generatedContent []byte, fixturePath string) error {
	got := generatedContent
	// Read the fixture content
	want, err := os.ReadFile(fixturePath)
	if err != nil {
		return fmt.Errorf("failed to read fixture: %w", err)
	}

	if normalizeWhitespace(want) != normalizeWhitespace(got) {
		if diff := cmp.Diff(string(want), string(got)); diff != "" {
			fmt.Println("Diff for file ", fixturePath, " (-want +got)")
			fmt.Printf(diff)
			return fmt.Errorf("generated content does not match fixture for file %s, check above for rich diff", fixturePath)
		}

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
