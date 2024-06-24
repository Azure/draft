package prompts

import (
	"io"
	"testing"

	"github.com/Azure/draft/pkg/config"
)

func TestGetVariableDefaultValue(t *testing.T) {
	tests := []struct {
		testName  string
		variables []config.BuilderVar
		inputs    map[string]string
		want      string
	}{
		{
			testName: "basicLiteralExtractDefault",
			variables: []config.BuilderVar{
				{
					Name: "var1",
					Default: config.BuilderVarDefault{
						Value: "default-value-1",
					},
				},
				{
					Name: "var2",
					Default: config.BuilderVarDefault{
						Value: "default-value-2",
					},
				},
			},
			inputs: map[string]string{},
			want:   "default-value-1",
		},
		{
			testName: "noDefaultIsEmptyString",
			variables: []config.BuilderVar{
				{
					Name: "var1",
				},
			},
			inputs: map[string]string{},
			want:   "",
		},
		{
			testName: "referenceTakesPrecedenceOverLiteral",
			variables: []config.BuilderVar{
				{
					Name: "var1",
					Default: config.BuilderVarDefault{
						ReferenceVar: "var2",
						Value:        "not-this-value",
					},
				},
			},
			inputs: map[string]string{
				"var2": "this-value",
			},
			want: "this-value",
		},
		{
			testName: "forwardReferencesAreIgnored",
			variables: []config.BuilderVar{
				{
					Name: "beforeVar",
					Default: config.BuilderVarDefault{
						ReferenceVar: "afterVar",
						Value:        "before-default-value",
					},
				},
				{
					Name: "afterVar",
					Default: config.BuilderVarDefault{
						Value: "not-this-value",
					},
				},
			},
			inputs: map[string]string{},
			want:   "before-default-value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := GetVariableDefaultValue(tt.variables[0], tt.inputs); got != tt.want {
				t.Errorf("GetVariableDefaultValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunStringPrompt(t *testing.T) {
	tests := []struct {
		testName         string
		prompt           config.BuilderVar
		userInputs       []string
		defaultValue     string
		want             string
		wantErr          bool
		mockDirNameValue string
	}{
		{
			testName: "basicPrompt",
			prompt: config.BuilderVar{
				Description: "var1 description",
			},
			userInputs:       []string{"value-1\n"},
			defaultValue:     "input",
			want:             "value-1",
			wantErr:          false,
			mockDirNameValue: "",
		},
		{
			testName: "promptWithDefault",
			prompt: config.BuilderVar{
				Description: "var1 description",
			},
			userInputs:       []string{"\n"},
			defaultValue:     "defaultValue",
			want:             "defaultValue",
			wantErr:          false,
			mockDirNameValue: "",
		},
		{
			testName: "appNameUsesDirName",
			prompt: config.BuilderVar{
				Name:        "APPNAME",
				Description: "app name",
			},
			userInputs:       []string{"\n"},
			defaultValue:     "currentdir",
			want:             "currentdir",
			wantErr:          false,
			mockDirNameValue: "currentdir",
		},
		{
			testName: "invalidAppName",
			prompt: config.BuilderVar{
				Name:        "APPNAME",
				Description: "app name",
			},
			userInputs:       []string{"--invalid-app-name\n"},
			defaultValue:     "defaultApp",
			want:             "",
			wantErr:          true,
			mockDirNameValue: "currentdir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Mock the getCurrentDirNameFunc for testing
			originalGetCurrentDirNameFunc := getCurrentDirNameFunc
			defer func() { getCurrentDirNameFunc = originalGetCurrentDirNameFunc }()
			getCurrentDirNameFunc = func() (string, error) {
				return tt.mockDirNameValue, nil
			}

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
			got, err := RunDefaultableStringPrompt(tt.defaultValue, tt.prompt, nil, inReader, nil)

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

func TestRunPromptsFromConfigWithSkipsIO(t *testing.T) {
	tests := []struct {
		testName     string
		config       config.DraftConfig
		userInputs   []string
		defaultValue string
		want         map[string]string
		wantErr      bool
	}{
		{
			testName: "onlyNoPrompt",
			config: config.DraftConfig{
				Variables: []config.BuilderVar{
					{
						Name: "var1",
						Default: config.BuilderVarDefault{
							IsPromptDisabled: true,
							Value:            "defaultValue",
						},
						Description: "var1 description",
					},
				},
			},
			userInputs: []string{""},
			want: map[string]string{
				"var1": "defaultValue",
			},
			wantErr: false,
		}, {
			testName: "twoPromptTwoNoPrompt",
			config: config.DraftConfig{
				Variables: []config.BuilderVar{
					{
						Name: "var1-no-prompt",
						Default: config.BuilderVarDefault{
							IsPromptDisabled: true,
							Value:            "defaultValueNoPrompt1",
						},
						Description: "var1 has IsPromptDisabled and should skip prompt and use default value",
					},
					{
						Name: "var2-default",
						Default: config.BuilderVarDefault{
							Value: "defaultValue2",
						},
						Description: "var2 has a default value and will receive an empty value, so it should use the default value",
					},
					{
						Name: "var3-no-prompt",
						Default: config.BuilderVarDefault{
							IsPromptDisabled: true,
							Value:            "defaultValueNoPrompt3",
						},
						Description: "var3 has IsPromptDisabled and should skip prompt and use default value",
					},
					{
						Name: "var4",
						Default: config.BuilderVarDefault{
							Value: "defaultValue4",
						},
						Description: "var4 has a default value, but has a value entered, so it should use the entered value",
					},
				},
			},
			userInputs: []string{"\n", "entered-value-for-4\n"},
			want: map[string]string{
				"var1-no-prompt": "defaultValueNoPrompt1",
				"var2-default":   "defaultValue2",
				"var3-no-prompt": "defaultValueNoPrompt3",
				"var4":           "entered-value-for-4",
			},
			wantErr: false,
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
			got, err := RunPromptsFromConfigWithSkipsIO(&tt.config, nil, inReader, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("TestRunPromptsFromConfigWithSkipsIO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for k, v := range got {
				wantVal := tt.want[k]
				if v != wantVal {
					t.Errorf("TestRunPromptsFromConfigWithSkipsIO()  inputs [%s]=%s, want %s", k, v, wantVal)
				}
			}
		})
	}
}
