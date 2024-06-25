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
		wantErr      bool
	}{
		{
			testName: "keepAllCustomInputs",
			draftConfig: DraftConfig{
				Variables: []BuilderVar{
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
			customInputs: map[string]string{
				"var1": "custom-value-1",
				"var2": "custom-value-2",
			},
			want: map[string]string{
				"var1": "custom-value-1",
				"var2": "custom-value-2",
			},
		},
		{
			testName: "applyDefaultToEmptyCustomInputs",
			draftConfig: DraftConfig{
				Variables: []BuilderVar{
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
			customInputs: map[string]string{},
			want: map[string]string{
				"var1": "default-value-1",
				"var2": "default-value-2",
			},
		},
		{
			testName: "applyDefaultToPartialCustomInputs",
			draftConfig: DraftConfig{
				Variables: []BuilderVar{
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
			customInputs: map[string]string{
				"var1": "custom-value-1",
			},
			want: map[string]string{
				"var1": "custom-value-1",
				"var2": "default-value-2",
			},
		},
		{
			testName: "variablesHaveNoInputOrDefault",
			draftConfig: DraftConfig{
				Variables: []BuilderVar{
					{
						Name: "var1",
					},
					{
						Name: "var2",
					},
				},
			},
			customInputs: map[string]string{},
			want:         map[string]string{},
			wantErr:      true,
		},
		{
			testName: "getDefaultFromReferenceVarCustomInputs",
			draftConfig: DraftConfig{
				Variables: []BuilderVar{
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
							Value: "not-this-value",
						},
					},
				},
			},
			customInputs: map[string]string{
				"var2": "this-value",
			},
			want: map[string]string{
				"var1": "this-value",
				"var2": "this-value",
			},
		},
		{
			testName: "getDefaultFromReferenceVar",
			draftConfig: DraftConfig{
				Variables: []BuilderVar{
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
			customInputs: map[string]string{},
			want: map[string]string{
				"var1": "default-value-3",
				"var2": "default-value-3",
				"var3": "default-value-3",
			},
		},
		{
			testName: "cyclicalReferenceVarsDetected",
			draftConfig: DraftConfig{
				Variables: []BuilderVar{
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
			customInputs: map[string]string{},
			want:         map[string]string{},
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if err := tt.draftConfig.ApplyDefaultVariables(tt.customInputs); (err != nil) != tt.wantErr {
				t.Error(err)
			} else {
				for k, v := range tt.want {
					if tt.customInputs[k] != v {
						t.Errorf("got: %s, want: %s", tt.customInputs[k], v)
					}
				}
			}
		})
	}
}
