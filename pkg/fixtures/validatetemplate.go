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
		genWords := strings.Split(normalizeWhitespace(generatedContent), " ")
		fixtureWords := strings.Split(normalizeWhitespace(fixtureContent), " ")
		differingWords := []string{}
		for i, word := range genWords {
			if word != fixtureWords[i] {
				differingWords = append(differingWords, fmt.Sprintf("'%s' != '%s'", word, fixtureWords[i]))
				if len(differingWords) == 1 {
					fmt.Println("Generated Word | Fixture Word")
				}
				fmt.Printf("'%s' != '%s'\n", word, fixtureWords[i])
			}
		}

		return errors.New(fmt.Sprintf("generated content does not match fixture: %s", strings.Join(differingWords, ", ")))
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
