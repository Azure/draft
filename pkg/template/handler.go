package template

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	tmpl "text/template"

	"github.com/bfoley13/draft/pkg/config"
	"github.com/bfoley13/draft/pkg/templatewriter"
	"github.com/bfoley13/draft/template"
	"github.com/blang/semver/v4"
)

const draftConfigFile = "draft.yaml"

var templateConfigs map[string]*Template

type Template struct {
	Config         *config.DraftConfig
	src            string
	dest           string
	templateFiles  fs.FS
	templateWriter templatewriter.TemplateWriter
	version        string
}

func init() {
	templateConfigs = make(map[string]*Template)

	err := fs.WalkDir(template.Templates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.EqualFold(d.Name(), draftConfigFile) {
			return nil
		}

		draftConfig, err := config.NewConfigFromFS(template.Templates, path)
		if err != nil {
			return err
		}

		if draftConfig.TemplateName == "" {
			return nil
		}

		if _, ok := templateConfigs[strings.ToLower(draftConfig.TemplateName)]; ok {
			return fmt.Errorf("duplicate template name: %s", draftConfig.TemplateName)
		}

		newTemplate := &Template{
			Config:        draftConfig,
			src:           filepath.Dir(path),
			templateFiles: template.Templates,
		}

		templateConfigs[strings.ToLower(draftConfig.TemplateName)] = newTemplate
		return nil
	})

	if err != nil {
		log.Fatalf("failed to init templates: %s", err.Error())
	}
}

func GetTemplate(name, version, dest string, templateWriter templatewriter.TemplateWriter) (*Template, error) {
	template, ok := templateConfigs[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("template %s not found", name)
	}

	template.dest = dest
	template.templateWriter = templateWriter
	if !isValidVersion(template.Config.Versions, version) {
		return nil, fmt.Errorf("template %s version %s not supported", name, version)
	}

	template.version = version

	return template, nil
}

func isValidVersion(versionRange, version string) bool {
	v, err := semver.Parse(version)
	if err != nil {
		return false
	}

	expectedRange, err := semver.ParseRange(versionRange)
	if err != nil {
		return false
	}

	return expectedRange(v)
}

func (t *Template) CreateTemplates() error {
	if err := t.validate(); err != nil {
		return err
	}

	if err := t.Config.ApplyDefaultVariablesForVersion(t.version); err != nil {
		return fmt.Errorf("create workflow files: %w", err)
	}

	if err := generateTemplate(t); err != nil {
		return err
	}

	return nil
}

func generateTemplate(template *Template) error {
	err := fs.WalkDir(template.GetTemplates(), template.GetSource(), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if strings.EqualFold(d.Name(), draftConfigFile) {
			return nil
		}

		if err := writeTemplate(template, path); err != nil {
			return err
		}

		return nil
	})

	return err
}

func writeTemplate(draftTemplate *Template, inputFile string) error {
	file, err := fs.ReadFile(draftTemplate.GetTemplates(), inputFile)
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

	if err = draftTemplate.GetTemplateWriter().WriteFile(fmt.Sprintf("%s/%s", draftTemplate.GetDestination(), filepath.Base(inputFile)), buf.Bytes()); err != nil {
		return err
	}

	return nil
}

func (t *Template) GetTemplateWriter() templatewriter.TemplateWriter {
	return t.templateWriter
}

func (t *Template) GetSource() string {
	return t.src
}

func (t *Template) GetTemplates() fs.FS {
	return t.templateFiles
}

func (t *Template) GetDestination() string {
	return t.dest
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

func (t *Template) IncludeInTemplateVersion(versionRange string) bool {
	v, err := semver.Parse(t.version)
	if err != nil {
		return false
	}

	expectedRange, err := semver.ParseRange(versionRange)
	if err != nil {
		return false
	}

	return expectedRange(v)
}
