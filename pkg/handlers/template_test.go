package handlers

import (
	"reflect"
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/stretchr/testify/assert"
)

func TestDeepCopy(t *testing.T) {
	// This will fail on adding a new field to the undelying structs that arent handled in DeepCopy
	testTemplate, err := GetTemplate("deployment-manifests", "0.0.1", ".", &writers.FileMapWriter{})
	assert.Nil(t, err)

	deepCopy := testTemplate.DeepCopy()

	assert.True(t, reflect.DeepEqual(deepCopy, testTemplate))
}
