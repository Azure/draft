package templatetests

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Azure/draft/pkg/fixtures"
	"github.com/Azure/draft/pkg/handlers"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/stretchr/testify/assert"
)

type TestInput struct {
	Name             string
	TemplateName     string
	FixturesBaseDir  string
	Version          string
	Dest             string
	TemplateWriter   *writers.FileMapWriter
	VarMap           map[string]string
	FileNameOverride map[string]string
	ExpectedErr      error
	Validators       map[string]func(string) error
	Transformers     map[string]func(string) (any, error)

	UseBaseFixtureWithFileNameOverride bool
	GenerateBaseTemplate               bool
}

func RunTemplateTest(t *testing.T, testInput TestInput) {
	t.Run(testInput.Name, func(t *testing.T) {
		template, err := handlers.GetTemplate(testInput.TemplateName, testInput.Version, testInput.Dest, testInput.TemplateWriter)
		assert.Nil(t, err)
		assert.NotNil(t, template)

		for k, v := range testInput.VarMap {
			template.Config.SetVariable(k, v)
		}

		for k, v := range testInput.Validators {
			template.Config.SetVariableValidator(k, v)
		}

		for k, v := range testInput.Transformers {
			template.Config.SetVariableTransformer(k, v)
		}

		overrideReverseLookup := make(map[string]string)
		for k, v := range testInput.FileNameOverride {
			template.Config.SetFileNameOverride(k, v)
			overrideReverseLookup[v] = k
		}

		err = template.Generate()
		if testInput.ExpectedErr != nil {
			if err == nil {
				t.Errorf("expected error %v, got nil", testInput.ExpectedErr)
				return
			}
			assert.True(t, strings.Contains(err.Error(), testInput.ExpectedErr.Error()))
			return
		}
		assert.Nil(t, err)

		for k, v := range testInput.TemplateWriter.FileMap {
			if testInput.GenerateBaseTemplate {
				err = os.MkdirAll(testInput.FixturesBaseDir, os.ModePerm)
				assert.Nil(t, err, "error creating base dir for new template fixture")
				err = os.WriteFile(fmt.Sprintf("%s/%s", testInput.FixturesBaseDir, k), []byte(v), os.ModePerm)
				assert.Nil(t, err, "error writing new template fixture")
				// skip the file validation checks
				continue
			}

			fileName := k
			if overrideFile, ok := overrideReverseLookup[filepath.Base(k)]; ok && testInput.UseBaseFixtureWithFileNameOverride {
				fileName = strings.Replace(fileName, filepath.Base(k), overrideFile, 1)
			}

			err = fixtures.ValidateContentAgainstFixture(v, fmt.Sprintf("%s/%s", testInput.FixturesBaseDir, fileName))
			assert.Nil(t, err)
		}
	})
}

func AlwaysFailingValidator(value string) error {
	return fmt.Errorf("this is a failing validator")
}

func AlwaysFailingTransformer(value string) (any, error) {
	return "", fmt.Errorf("this is a failing transformer")
}

func K8sLabelValidator(value string) error {
	labelRegex, err := regexp.Compile("^((A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$")
	if err != nil {
		return err
	}
	if !labelRegex.MatchString(value) {
		return fmt.Errorf("invalid label: %s", value)
	}
	return nil
}
