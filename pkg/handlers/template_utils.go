package handlers

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/template"
	"github.com/blang/semver/v4"
	log "github.com/sirupsen/logrus"
)

var templateConfigs map[string]*Template

func init() {
	if err := loadTemplates(); err != nil {
		log.Fatalf("failed to init templates: %s", err.Error())
	}
}

// GetTemplates returns all templates
func GetTemplates() map[string]*Template {
	return templateConfigs
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
func IsValidVersion(versionRange, version string) bool {
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

func sanatizeTemplateSrcDir(src string) string {
	srcDir := filepath.Dir(src)

	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(srcDir, "\\", "/")
	}

	return srcDir
}
