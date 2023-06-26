package defaults

import (
	"fmt"
	"strings"

	"github.com/Azure/draft/pkg/reporeader"
	log "github.com/sirupsen/logrus"
)

type GradleExtractor struct {
}

// GetName implements reporeader.VariableExtractor
func (*GradleExtractor) GetName() string {
	return "gradle"
}

// MatchesLanguage implements reporeader.VariableExtractor
func (*GradleExtractor) MatchesLanguage(lowerlang string) bool {
	return lowerlang == "gradle" || lowerlang == "gradlew"
}

// ReadDefaults implements reporeader.VariableExtractor
func (*GradleExtractor) ReadDefaults(r reporeader.RepoReader) (map[string]string, error) {
	extractedValues := make(map[string]string)
	files, err := r.FindFiles(".", []string{"*.gradle"}, 2)
	if err != nil {
		return nil, fmt.Errorf("error finding gradle files: %v", err)
	}
	if len(files) > 0 {
		f, err := r.ReadFile(files[0])
		if err != nil {
			log.Warn("Unable to read build.gradle, skipping detection")
			return nil, nil
		}
		content := string(f)
		// this separator is used to split the line from build.gradle ex: sourceCompatibility = '1.8'
		// output will be ['sourceCompatibility', '1.8'] or ["sourceCompatibility", "1.8"]
		separator := func(c rune) bool {
			return c == ' ' || c == '=' || c == '\n' || c == '\r' || c == '\t' || c == '{' || c == '}' || c == '[' || c == ']' || c == '-'
		}
		// this func takes care of removing the single or double quotes from split array output
		cutset := func(c rune) bool { return c == '\'' || c == '"' }
		if strings.Contains(content, "sourceCompatibility") || strings.Contains(content, "targetCompatibility") || strings.Contains(content, "server.port") {
			stringAfterSplit := strings.FieldsFunc(content, separator)
			for i := 0; i < len(stringAfterSplit); i++ {
				if stringAfterSplit[i] == "sourceCompatibility" {
					detectedVersion := strings.TrimFunc(stringAfterSplit[i+1], cutset)
					detectedVersion = detectedVersion + "-jre"
					extractedValues["VERSION"] = detectedVersion
				} else if stringAfterSplit[i] == "targetCompatibility" {
					detectedBuilderVersion := strings.TrimFunc(stringAfterSplit[i+1], cutset)
					detectedBuilderVersion = "jdk" + detectedBuilderVersion
					extractedValues["BUILDERVERSION"] = detectedBuilderVersion
				} else if stringAfterSplit[i] == "server.port" {
					detectedPort := strings.TrimFunc(stringAfterSplit[i+1], cutset)
					extractedValues["PORT"] = detectedPort
				}
			}
		}
	}

	return extractedValues, nil
}

var _ reporeader.VariableExtractor = &GradleExtractor{}
