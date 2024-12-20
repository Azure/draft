package config

import (
	"testing"
)

func TestApplyDefaultVariables(t *testing.T) {
	tests := []struct {
		testName     string
		draftConfig  DraftConfig
		customInputs map[string]string
		want         map[string]string
		wantErrMsg   string
	}{
		{
			testName: "keepAllCustomInputs",
			draftConfig: DraftConfig{
				Variables: []*BuilderVar{
					{
						Name:  "var1",
						Value: "custom-value-1",
						Default: BuilderVarDefault{
							Value: "default-value-1",
						},
					},
					{
						Name:  "var2",
						Value: "custom-value-2",
						Default: BuilderVarDefault{
							Value: "default-value-2",
						},
					},
				},
			},
			want: map[string]string{
				"var1": "custom-value-1",
				"var2": "custom-value-2",
			},
		},
		{
			testName: "applyDefaultToEmptyCustomInputs",
			draftConfig: DraftConfig{
				Variables: []*BuilderVar{
					{
						Name: "var1",
						Default: BuilderVarDefault{
							Value: "default-value-1",
						},
					},
					{
						Name: "var2",
						Default: BuilderVarDefault{
							Value: "default-value-2",
						},
					},
				},
			},
			want: map[string]string{
				"var1": "default-value-1",
				"var2": "default-value-2",
			},
		},
		{
			testName: "applyDefaultToPartialCustomInputs",
			draftConfig: DraftConfig{
				Variables: []*BuilderVar{
					{
						Name:  "var1",
						Value: "custom-value-1",
						Default: BuilderVarDefault{
							Value: "default-value-1",
						},
					},
					{
						Name: "var2",
						Default: BuilderVarDefault{
							Value: "default-value-2",
						},
					},
				},
			},
			want: map[string]string{
				"var1": "custom-value-1",
				"var2": "default-value-2",
			},
		},
		{
			testName: "variablesHaveNoInputOrDefault",
			draftConfig: DraftConfig{
				Variables: []*BuilderVar{
					{
						Name: "var1",
					},
					{
						Name: "var2",
					},
				},
			},
			want:       map[string]string{},
			wantErrMsg: "variable var1 has no default value",
		},
		{
			testName: "getDefaultFromReferenceVarCustomInputs",
			draftConfig: DraftConfig{
				Variables: []*BuilderVar{
					{
						Name: "var1",
						Default: BuilderVarDefault{
							ReferenceVar: "var2",
							Value:        "not-this-value",
						},
					},
					{
						Name:  "var2",
						Value: "this-value",
						Default: BuilderVarDefault{
							Value: "not-this-value",
						},
					},
				},
			},
			want: map[string]string{
				"var1": "this-value",
				"var2": "this-value",
			},
		},
		{
			testName: "getDefaultFromReferenceVar",
			draftConfig: DraftConfig{
				Variables: []*BuilderVar{
					{
						Name: "var1",
						Default: BuilderVarDefault{
							ReferenceVar: "var2",
							Value:        "not-this-value",
						},
					},
					{
						Name: "var2",
						Default: BuilderVarDefault{
							ReferenceVar: "var3",
							Value:        "not-this-value",
						},
					},
					{
						Name: "var3",
						Default: BuilderVarDefault{
							Value: "default-value-3",
						},
					},
				},
			},
			want: map[string]string{
				"var1": "default-value-3",
				"var2": "default-value-3",
				"var3": "default-value-3",
			},
		},
		{
			testName: "cyclicalReferenceVarsDetected",
			draftConfig: DraftConfig{
				Variables: []*BuilderVar{
					{
						Name: "var1",
						Default: BuilderVarDefault{
							ReferenceVar: "var2",
						},
					},
					{
						Name: "var2",
						Default: BuilderVarDefault{
							ReferenceVar: "var1",
						},
					},
				},
			},
			want:       map[string]string{},
			wantErrMsg: "apply default variables: cyclical reference detected",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if err := tt.draftConfig.ApplyDefaultVariables(); err != nil && err.Error() != tt.wantErrMsg {
				t.Error(err)
			} else {
				for k, v := range tt.want {
					variable, err := tt.draftConfig.GetVariable(k)
					if err != nil {
						t.Error(err)
					}

					if variable.Value != v {
						t.Errorf("got: %s, want: %s, for test: %s", variable.Value, v, tt.testName)
					}
				}
			}
		})
	}
}

func TestApplyDefaultVariablesForVersion(t *testing.T) {
	tests := []struct {
		testName     string
		version      string
		draftConfig  DraftConfig
		customInputs map[string]string
		want         map[string]string
		wantErrMsg   string
	}{
		{
			testName: "excludeOutOfVersionRangeVariables",
			version:  "0.0.1",
			draftConfig: DraftConfig{
				Versions: []string{"0.0.1", "0.0.2"},
				Variables: []*BuilderVar{
					{
						Name:     "var1",
						Value:    "",
						Versions: ">=0.0.1",
						Default: BuilderVarDefault{
							Value: "default-value-1",
						},
					},
					{
						Name:     "var2",
						Value:    "custom-value-2",
						Versions: ">=0.0.2",
						Default: BuilderVarDefault{
							Value: "default-value-2",
						},
					},
				},
			},
			want: map[string]string{
				"var1": "default-value-1",
			},
		},
		{
			testName: "emptyInputVersion",
			version:  "",
			draftConfig: DraftConfig{
				Versions: []string{"0.0.1", "0.0.2"},
				Variables: []*BuilderVar{
					{
						Name:     "var1",
						Value:    "",
						Versions: ">=0.0.1",
						Default: BuilderVarDefault{
							Value: "default-value-1",
						},
					},
					{
						Name:     "var2",
						Value:    "",
						Versions: ">=0.0.2",
						Default: BuilderVarDefault{
							Value: "default-value-2",
						},
					},
				},
			},
			wantErrMsg: "invalid version: Version string empty",
		},
		{
			testName: "inputVersionOutOfRange",
			version:  "0.0.3",
			draftConfig: DraftConfig{
				Versions: []string{"0.0.1", "0.0.2"},
				Variables: []*BuilderVar{
					{
						Name:     "var1",
						Value:    "",
						Versions: ">=0.0.1",
						Default: BuilderVarDefault{
							Value: "default-value-1",
						},
					},
					{
						Name:     "var2",
						Value:    "",
						Versions: ">=0.0.2",
						Default: BuilderVarDefault{
							Value: "default-value-2",
						},
					},
				},
			},
			wantErrMsg: "requested version outside of valid versions: 0.0.3",
		},
		{
			testName: "overwriteDevfaultValue",
			version:  "0.0.2",
			draftConfig: DraftConfig{
				Versions: []string{"0.0.1", "0.0.2"},
				Variables: []*BuilderVar{
					{
						Name:     "var1",
						Value:    "custom-value-1",
						Versions: ">=0.0.1",
						Default: BuilderVarDefault{
							Value: "default-value-1",
						},
					},
					{
						Name:     "var2",
						Value:    "custom-value-2",
						Versions: ">=0.0.2",
						Default: BuilderVarDefault{
							Value: "default-value-2",
						},
					},
				},
			},
			want: map[string]string{
				"var1": "custom-value-1",
				"var2": "custom-value-2",
			},
		},
		{
			testName: "referenceVarOverwrite",
			version:  "0.0.2",
			draftConfig: DraftConfig{
				Versions: []string{"0.0.1", "0.0.2"},
				Variables: []*BuilderVar{
					{
						Name:     "var1",
						Value:    "",
						Versions: ">=0.0.1",
						Default: BuilderVarDefault{
							ReferenceVar: "var2",
						},
					},
					{
						Name:     "var2",
						Value:    "custom-value-2",
						Versions: ">=0.0.2",
						Default: BuilderVarDefault{
							Value: "default-value-2",
						},
					},
				},
			},
			want: map[string]string{
				"var1": "custom-value-2",
				"var2": "custom-value-2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if err := tt.draftConfig.ApplyDefaultVariablesForVersion(tt.version); err != nil && err.Error() != tt.wantErrMsg {
				t.Error(err)
			} else {
				for k, v := range tt.want {
					variable, err := tt.draftConfig.GetVariable(k)
					if err != nil {
						t.Error(err)
					}

					if variable.Value != v {
						t.Errorf("got: %s, want: %s, for test: %s", variable.Value, v, tt.testName)
					}
				}
			}
		})
	}
}
