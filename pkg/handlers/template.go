package handlers

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	tmpl "text/template"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/handlers/variableextractors/defaults"
	"github.com/Azure/draft/pkg/reporeader"
	"github.com/Azure/draft/pkg/templatewriter"
	log "github.com/sirupsen/logrus"
)

type Template struct {
	Config *config.DraftConfig

	templateFiles  fs.FS
	templateWriter templatewriter.TemplateWriter
	src            string
	dest           string
	version        string
}

// GetTemplate returns a template by name, version, and destination
func GetTemplate(name, version, dest string, templateWriter templatewriter.TemplateWriter) (*Template, error) {
	template, ok := templateConfigs[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	template = template.DeepCopy()

	if version == "" {
		version = template.Config.DefaultVersion
		log.Println("version not provided, using default version: ", version)
	}

	if !IsValidVersion(template.Config.Versions, version) {
		return nil, fmt.Errorf("invalid version: %s", version)
	}

	if dest == "" {
		dest = "."
		log.Println("destination not provided, using current directory")
	}

	if _, err := filepath.Abs(dest); err != nil {
		return nil, fmt.Errorf("invalid destination: %s", dest)
	}

	template.dest = dest
	template.version = version
	template.templateWriter = templateWriter

	return template, nil
}

func (t *Template) Generate() error {
	if err := t.validate(); err != nil {
		log.Printf("template validation failed: %s", err.Error())
		return err
	}

	if err := t.Config.ApplyDefaultVariablesForVersion(t.version); err != nil {
		return fmt.Errorf("create workflow files: %w", err)
	}

	return generateTemplate(t)
}

func (t *Template) validate() error {
	if t == nil {
		return fmt.Errorf("template is nil")
	}

	if t.Config == nil {
		return fmt.Errorf("template draft config is nil")
	}

	if t.src == "" {
		return fmt.Errorf("template source is empty")
	}

	if t.dest == "" {
		return fmt.Errorf("template destination is empty")
	}

	if t.templateFiles == nil {
		return fmt.Errorf("template files is nil")
	}

	if t.version == "" {
		return fmt.Errorf("template version is empty")
	}

	return nil
}

func (t *Template) DeepCopy() *Template {
	return &Template{
		Config:         t.Config.DeepCopy(),
		templateFiles:  t.templateFiles,
		templateWriter: t.templateWriter,
		src:            t.src,
		dest:           t.dest,
		version:        t.version,
	}
}

func (l *Template) ExtractDefaults(lowerLang string, r reporeader.RepoReader) (map[string]string, error) {
	extractors := []reporeader.VariableExtractor{
		&defaults.PythonExtractor{},
		&defaults.GradleExtractor{},
	}
	extractedValues := make(map[string]string)
	if r == nil {
		log.Debugf("no repo reader provided, returning empty list of defaults")
		return extractedValues, nil
	}
	for _, extractor := range extractors {
		if extractor.MatchesLanguage(lowerLang) {
			newDefaults, err := extractor.ReadDefaults(r)
			if err != nil {
				return nil, fmt.Errorf("error reading defaults for language %s: %v", lowerLang, err)
			}
			for k, v := range newDefaults {
				if _, ok := extractedValues[k]; ok {
					log.Debugf("duplicate default %s for language %s with extractor %s", k, lowerLang, extractor.GetName())
				}
				extractedValues[k] = v
				log.Debugf("extracted default %s=%s with extractor:%s", k, v, extractor.GetName())
			}
		}
	}

	return extractedValues, nil
}

func generateTemplate(template *Template) error {
	err := fs.WalkDir(template.templateFiles, template.src, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return template.templateWriter.EnsureDirectory(strings.Replace(path, template.src, template.dest, 1))
		}

		if strings.EqualFold(d.Name(), "draft.yaml") {
			return nil
		}

		if err := writeTemplate(template, path); err != nil {
			return fmt.Errorf("failed to write template %s: %w", path, err)
		}

		return nil
	})

	return err
}

func writeTemplate(draftTemplate *Template, inputFile string) error {
	file, err := fs.ReadFile(draftTemplate.templateFiles, inputFile)
	if err != nil {
		return err
	}

	// Parse the template file, missingkey=error ensures an error will be returned if any variable is missing during template execution.
	tmpl, err := tmpl.New("template").Option("missingkey=error").Parse(string(file))
	if err != nil {
		return err
	}

	// Execute the template with variableMap
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, draftTemplate)
	if err != nil {
		return err
	}

	if err = draftTemplate.templateWriter.WriteFile(getOutputFileName(draftTemplate, inputFile), buf.Bytes()); err != nil {
		return err
	}

	return nil
}

func getOutputFileName(draftTemplate *Template, inputFile string) string {
	outputName := filepath.Clean(strings.Replace(inputFile, draftTemplate.src, draftTemplate.dest, 1))

	fileName := filepath.Base(inputFile)
	if overrideName, ok := draftTemplate.Config.FileNameOverrideMap[fileName]; ok {
		return strings.Replace(outputName, fileName, overrideName, 1)
	}

	return outputName
}
