package guardrails

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	Default_Manifest                        = "Default_Manifest.yaml"
	ContainerAllowedImages_Error_Manifest   = "CAI_Error_Manifest.yaml"
	ContainerAllowedImages_Success_Manifest = "CAI_Success_Manifest.yaml"
)

func TestValidateGuardrailsConstraint_Default(t *testing.T) {
	df := Default_Manifest
	err := ValidateGuardrailsConstraint(df)
	assert.Nil(t, err)
}

// thbarnes: working on error case(s); investigate if default can just be used for all success
func TestValidateGuardrailsConstraint_CAI(t *testing.T) {
	//df := ContainerAllowedImages_Error_Manifest
	//err := ValidateGuardrailsConstraint(df)
	//assert.NotNil(t, err)
	// get more specific once we're certain parameters are set properly in deployment

	df := ContainerAllowedImages_Success_Manifest
	err := ValidateGuardrailsConstraint(df)
	assert.Nil(t, err)
}
