package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootInitConfig(t *testing.T) {
	initConfig()

	cfgFile := "./../test/templatesgit ch/config.yaml"
	initConfig()

	assert.True(t, cfgFile != "")
}
