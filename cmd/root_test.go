package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootInitConfig(t *testing.T) {
	initConfig()

	cfgFile := "./../test/template/config.yaml"
	initConfig()

	assert.True(t, cfgFile != "")
}
