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

//go:embed deployments/*
var deployments embed.FS

func ValidateContentAgainstFixture(generatedContent []byte, fixturePath string) error {
	fullFixturePath := fmt.Sprintf("%s", fixturePath)

	// Read the fixture content
	fixtureContent, err := os.ReadFile(fullFixturePath)
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
