package prompts

import (
	"testing"

	"github.com/Azure/draft/pkg/config"
)

func TestGetVariableDefaultValue(t *testing.T) {
	tests := []struct {
		testName         string
		variableName     string
		variableDefaults []config.BuilderVarDefault
		inputs           map[string]string
		want             string
	}{
		{
			testName:     "basicLiteralExtractDefault",
			variableName: "var1",
			variableDefaults: []config.BuilderVarDefault{
				{
					Name:  "var1",
					Value: "default-value-1",
				},
				{
					Name:  "var2",
					Value: "default-value-2",
				},
			},
			inputs: map[string]string{},
			want:   "default-value-1",
		},
		{
			testName:         "noDefaultIsEmptyString",
			variableName:     "var1",
			variableDefaults: []config.BuilderVarDefault{},
			inputs:           map[string]string{},
			want:             "",
		},
		{
			testName:     "referenceTakesPrecedenceOverLiteral",
			variableName: "var1",
			variableDefaults: []config.BuilderVarDefault{
				{
					Name:         "var1",
					Value:        "not-this-value",
					ReferenceVar: "var2",
				},
			},
			inputs: map[string]string{
				"var2": "this-value",
			},
			want: "this-value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := GetVariableDefaultValue(tt.variableName, tt.variableDefaults, tt.inputs); got != tt.want {
				t.Errorf("GetVariableDefaultValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
