package transformers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTransformer(t *testing.T) {
	assert.NotNil(t, GetTransformer("NonExistentKind"))
}

func TestDefaultTransformer(t *testing.T) {
	res, err := DefaultTransformer("test")
	assert.Nil(t, err)
	assert.Equal(t, "test", res)
}
