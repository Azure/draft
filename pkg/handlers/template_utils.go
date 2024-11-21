package handlers

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/template"
	"github.com/blang/semver/v4"
	log "github.com/sirupsen/logrus"
)

var templateConfigs map[string]*Template

type TemplateType string

func (t TemplateType) String() string {
	return string(t)
}

const (
	TemplateTypeDeployment TemplateType = "deployment"
	TemplateTypeDockerfile TemplateType = "dockerfile"
	TemplateTypeManifests  TemplateType = "manifest"
	TemplateTypeWorkflow   TemplateType = "workflow"
)

func init() {
	if err := loadTemplates(); err != nil {
		log.Fatalf("failed to init templates: %s", err.Error())
	}
}

// GetTemplates returns all templates
func GetTemplates() map[string]*Template {
	return templateConfigs
}

func GetTemplatesByType(templateType TemplateType) map[string]*Template {
	templates := make(map[string]*Template)
	for name, template := range templateConfigs {
		if template.Config.Type == templateType.String() {
			templates[name] = template
		}
	}
	return templates
}

func IsValidTemplate(templateName string) bool {
	_, ok := templateConfigs[strings.ToLower(templateName)]
	return ok
}

func loadTemplates() error {
	templateConfigs = make(map[string]*Template)
	return fs.WalkDir(template.Templates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.EqualFold(d.Name(), "draft.yaml") {
			return nil
		}

		draftConfig, err := config.NewConfigFromFS(template.Templates, path)
		if err != nil {
			return err
		}

		if _, ok := templateConfigs[strings.ToLower(draftConfig.TemplateName)]; ok {
			return fmt.Errorf("duplicate template name: %s", draftConfig.TemplateName)
		}

		newTemplate := &Template{
			Config:        draftConfig,
			src:           sanatizeTemplateSrcDir(path),
			templateFiles: template.Templates,
		}

		templateConfigs[strings.ToLower(draftConfig.TemplateName)] = newTemplate
		return nil
	})
}

// IsValidVersion checks if a version is valid for a given version range
func IsValidVersion(versions []string, version string) bool {
	_, err := semver.Parse(version)
	if err != nil {
		return false
	}

	return slices.Contains(versions, version)
}

func sanatizeTemplateSrcDir(src string) string {
	srcDir := filepath.Dir(src)

	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(srcDir, "\\", "/")
	}

	return srcDir
}
