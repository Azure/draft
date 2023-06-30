package defaults

import (
	"fmt"
	"strings"

	"github.com/Azure/draft/pkg/reporeader"
	log "github.com/sirupsen/logrus"
)

const SOURCE_COMPATIBILITY = "sourceCompatibility"
const TARGET_COMPATIBILITY = "targetCompatibility"
const SERVER_PORT = "server.port"
const GRADLE_FILE_FORMAT = "*.gradle"

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
	separatorsSet := createSeparatorsSet()
	cutSet := createCutSet()
	extractedValues := make(map[string]string)
	files, err := r.FindFiles(".", []string{GRADLE_FILE_FORMAT}, 2)
	if err != nil {
		return nil, fmt.Errorf("error finding gradle files: %v", err)
	}
	if len(files) > 0 {
		f, err := r.ReadFile(files[0])
		if err != nil {
			log.Warn("Unable to read gradle file, skipping detection")
			return nil, nil
		}
		content := string(f)
		// this separator is used to split the line from build.gradle ex: sourceCompatibility = '1.8'
		// output will be ['sourceCompatibility', '1.8'] or ["sourceCompatibility", "1.8"]
		separatorFunc := func(c rune) bool {
			return separatorsSet.Contains(c)
		}
		// this func takes care of removing the single or double quotes from split array output
		cutset := func(c rune) bool { return cutSet.Contains(c) }
		if strings.Contains(content, SOURCE_COMPATIBILITY) || strings.Contains(content, TARGET_COMPATIBILITY) || strings.Contains(content, SERVER_PORT) {
			stringAfterSplit := strings.FieldsFunc(content, separatorFunc)
			for i, s := range stringAfterSplit {
				if i+1 >= len(stringAfterSplit) {
					break
				}
				if s == SOURCE_COMPATIBILITY {
					detectedVersion := strings.TrimFunc(stringAfterSplit[i+1], cutset)
					detectedVersion = detectedVersion + "-jre"
					extractedValues["VERSION"] = detectedVersion
				} else if s == TARGET_COMPATIBILITY {
					detectedBuilderVersion := strings.TrimFunc(stringAfterSplit[i+1], cutset)
					detectedBuilderVersion = "jdk" + detectedBuilderVersion
					extractedValues["BUILDERVERSION"] = detectedBuilderVersion
				} else if s == SERVER_PORT {
					detectedPort := strings.TrimFunc(stringAfterSplit[i+1], cutset)
					extractedValues["PORT"] = detectedPort
				}
			}
		}
	}

	return extractedValues, nil
}

func createSeparatorsSet() Set {
	separatorsSet := NewSet()
	separatorsSet.Add(' ')
	separatorsSet.Add('=')
	separatorsSet.Add('\n')
	separatorsSet.Add('\r')
	separatorsSet.Add('\t')
	separatorsSet.Add('{')
	separatorsSet.Add('}')
	separatorsSet.Add('[')
	separatorsSet.Add(']')
	separatorsSet.Add('-')
	separatorsSet.Add(':')
	return separatorsSet
}

func createCutSet() Set {
	cutSet := NewSet()
	cutSet.Add('\'')
	cutSet.Add('"')
	return cutSet
}

type Set map[interface{}]struct{}

func NewSet() Set {
	return make(Set)
}

func (s Set) Add(item interface{}) {
	s[item] = struct{}{}
}

func (s Set) Contains(item interface{}) bool {
	_, ok := s[item]
	return ok
}

var _ reporeader.VariableExtractor = &GradleExtractor{}
