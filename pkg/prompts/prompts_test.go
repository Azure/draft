package prompts

import (
	"io"
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
		}, {
			testName:     "forwardReferencesAreIgnored",
			variableName: "beforeVar",
			variableDefaults: []config.BuilderVarDefault{
				{
					Name:         "beforeVar",
					Value:        "before-default-value",
					ReferenceVar: "afterVar",
				}, {
					Name:  "afterVar",
					Value: "not-this-value",
				},
			},
			inputs: map[string]string{},
			want:   "before-default-value",
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

func TestRunStringPrompt(t *testing.T) {
	tests := []struct {
		testName     string
		prompt       config.BuilderVar
		userInputs   []string
		defaultValue string
		want         string
		wantErr      bool
	}{
		{
			testName: "basicPrompt",
			prompt: config.BuilderVar{
				Name:        "var1",
				Description: "var1 description",
			},
			userInputs:   []string{"value-1\n"},
			defaultValue: "input",
			want:         "value-1",
			wantErr:      false,
		},
		{
			testName: "promptWithDefault",
			prompt: config.BuilderVar{
				Name:        "var1",
				Description: "var1 description",
			},
			userInputs:   []string{"\n"},
			defaultValue: "defaultValue",
			want:         "defaultValue",
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			inReader, inWriter := io.Pipe()

			go func() {
				for _, input := range tt.userInputs {
					_, err := inWriter.Write([]byte(input))
					if err != nil {
						t.Errorf("Error writing to inWriter: %v", err)
					}
				}
				err := inWriter.Close()
				if err != nil {
					t.Errorf("Error closing inWriter: %v", err)
				}
			}()
			got, err := RunDefaultableStringPrompt(tt.prompt, tt.defaultValue, nil, inReader, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunDefaultableStringPrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RunDefaultableStringPrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
