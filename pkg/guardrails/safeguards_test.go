package safeguards

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
)

const (
	Default_Manifest = "Default_Manifest.yaml"

	Constraint_CAI = "container-allowed-images"
	Constraint_CEP = "container-enforce-probes"
	Constraint_CRP = "container-resource-limits"
	Constraint_NUP = "no-unauthenticated-pulls"
	Constraint_PDB = "pod-disruption-budgets"
	Constraint_PEA = "pod-enforce-antiaffinity"
	Constraint_RT  = "restricted-taints"
	Constraint_USS = "unique-service-selectors"
)

func generateTestCases(constraintName string) []string {
	words := strings.Split(constraintName, "-")

	w1 := strings.ToUpper(strconv.Itoa(int(words[0][0])))
	w2 := strings.ToUpper(strconv.Itoa(int(words[1][0])))
	w3 := strings.ToUpper(strconv.Itoa(int(words[2][0])))

	abbreviation := w1 + w2 + w3

	testNameError := abbreviation + "_Error_Manifest"
	testNameSuccess := abbreviation + "_Success_Manifest"

	return []string{testNameError, testNameSuccess}
}

func TestValidateSafeguardsConstraint_Default(t *testing.T) {
	df := Default_Manifest
	err := ValidateDeployment(df, "")
	assert.Nil(t, err)
}

// thbarnes: working on error case(s); investigate if default can just be used for all success
func TestValidateSafeguardsConstraint_CAI(t *testing.T) {
	c := Constraint_CAI
	testCases := generateTestCases(c)
	for i, df := range testCases {
		if i == 0 { // error
			err := ValidateDeployment(df, c)
			assert.NotNil(t, err)
		} else { // success
			err := ValidateDeployment(df, "")
			assert.Nil(t, err)
		}
	}
}
