package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValidator(t *testing.T) {
	assert.NotNil(t, GetValidator("NonExistentKind"))
}

func TestDefaultValidator(t *testing.T) {
	assert.Nil(t, DefaultValidator("test"))
}
