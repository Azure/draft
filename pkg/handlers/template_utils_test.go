package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplate(t *testing.T) {
	loadedTemplates := GetTemplates()
	assert.Positive(t, len(loadedTemplates))
}

func TestLoadTemplates(t *testing.T) {
	templateConfigs = nil
	err := loadTemplates()
	assert.Nil(t, err)
	loadedTemplates := GetTemplates()
	assert.Positive(t, len(loadedTemplates))
}
