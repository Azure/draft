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

func GetValidTemplateVersions(templateName string) ([]string, error) {
	template, ok := templateConfigs[strings.ToLower(templateName)]
	if !ok {
		return nil, fmt.Errorf("template %s not found", templateName)
	}

	validRanges := strings.Split(template.Config.Versions, "||")
	if len(validRanges) == 0 {
		return nil, fmt.Errorf("no versions found for template %s", templateName)
	}

	for _, validRange := range validRanges {
		versions := strings.Split(strings.Trim(validRange, " "), ",")
		allowedVersions := make([]semver.Version, 0)
		lessThanVersions := make([]semver.Version, 0)
		greaterthanVersions := make([]semver.Version, 0)
		for _, version := range versions {
			versionCondition := getVersionCondition(version)
			semVersion, err := semver.Parse(strings.Trim(version, versionCondition))
			if err != nil {
				return nil, err
			}

			switch versionCondition {
			case "<=":
				lessThanVersions = append(lessThanVersions, semVersion)
			case ">=":
				greaterthanVersions = append(greaterthanVersions, semVersion)
			case "<":
				lessThanVersions = append(lessThanVersions, semVersion.PrevPatch())
			case ">":
				greaterthanVersions = append(greaterthanVersions, semVersion.NextPatch())
			case "=", "==":
				allowedVersions = append(allowedVersions, semVersion)
			default:
				allowedVersions = append(allowedVersions, semVersion)
			}
		}

		for _, greaterThanVersion := range greaterthanVersions {
			hasBound := false
			var minVersionDiff *semver.Version
			for _, lessThanVersion := range lessThanVersions {
				if greaterThanVersion.GT(lessThanVersion) {
					continue
				}

				versionDiff, err := semver.Parse(fmt.Sprintf("%d.%d.%d", lessThanVersion.Major-greaterThanVersion.Major, lessThanVersion.Minor-greaterThanVersion.Minor, lessThanVersion.Patch-greaterThanVersion.Patch))
				if err != nil {
					return nil, fmt.Errorf("failed to parse version difference: %w", err)
				}

				if versionDiff.Major != 0 && versionDiff.Minor != 0 {
					continue
				}

				if minVersionDiff == nil {
					minVersionDiff = &versionDiff
				}

				if versionDiff.Patch < minVersionDiff.Patch {
					minVersionDiff = &versionDiff
				}
				hasBound = true
			}
			if !hasBound {
				return nil, fmt.Errorf("unbounded version range: %s", validRange)
			}

			for i := 0; i < minVersionDiff.Patch; i++ {


		}
	}

	return template.Config.Versions
}

func getVersionCondition(version string) string {
	if strings.HasPrefix(version, "<=") {
		return "<="
	} else if strings.HasPrefix(version, ">=") {
		return ">="
	} else if strings.HasPrefix(version, "==") {
		return "=="
	} else if strings.HasPrefix(version, "=") {
		return "="
	} else if strings.HasPrefix(version, ">") {
		return ">"
	} else if strings.HasPrefix(version, "<") {
		return "<"
	}

	return ""
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
