package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetVersionAtRuntime(t *testing.T) {
	vcsInfo := getVCSInfoFromRuntime()
	assert.Empty(t, vcsInfo)
}
