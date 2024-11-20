package config

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetValidTemplateVersions(t *testing.T) {
	tests := []struct {
		Version        string
		Valid          bool
		OutputVersions []string
	}{
		{
			Version:        "1.0.0",
			Valid:          true,
			OutputVersions: []string{"1.0.0"},
		},
		{
			Version: "<1.0.0",
			Valid:   false,
		},
		{
			Version:        ">=0.0.1 <=0.0.5 || >0.1.0 <0.1.5",
			Valid:          true,
			OutputVersions: []string{"0.0.1", "0.0.2", "0.0.3", "0.0.4", "0.0.5", "0.1.1", "0.1.2", "0.1.3", "0.1.4"},
		},
		{
			Version: ">=0.0.1 <=0.0.5 || >0.1.0 <0.2.5",
			Valid:   false,
		},
		{
			Version:        ">=0.0.1 <0.0.5",
			Valid:          true,
			OutputVersions: []string{"0.0.1", "0.0.2", "0.0.3", "0.0.4"},
		},
		{
			Version: ">=0.0.1 0.0.5",
			Valid:   false,
		},
		{
			Version: "<0.1.1",
			Valid:   false,
		},
		{
			Version:        ">=0.0.1 <=0.0.6 !0.0.4",
			Valid:          true,
			OutputVersions: []string{"0.0.1", "0.0.2", "0.0.3", "0.0.5", "0.0.6"},
		},
		{
			Version:        ">=0.0.1 <=0.0.6 !0.0.6",
			Valid:          true,
			OutputVersions: []string{"0.0.1", "0.0.2", "0.0.3", "0.0.4", "0.0.5"},
		},
		{
			Version:        ">0.0.1 >=0.0.2 <0.0.7 !0.0.6",
			Valid:          true,
			OutputVersions: []string{"0.0.2", "0.0.3", "0.0.4", "0.0.5"},
		},
		{
			Version:        ">0.0.1 >=0.0.2 <0.0.7 <=0.0.8 !0.0.6",
			Valid:          true,
			OutputVersions: []string{"0.0.2", "0.0.3", "0.0.4", "0.0.5", "0.0.7", "0.0.8"},
		},
		{
			Version:        ">0.0.1 <=0.0.5 || >=0.0.5 <0.0.9 !0.0.6",
			Valid:          true,
			OutputVersions: []string{"0.0.2", "0.0.3", "0.0.4", "0.0.5", "0.0.7", "0.0.8"},
		},
	}

	for _, test := range tests {
		versions, err := GetValidTemplateVersions(test.Version)
		if test.Valid {
			assert.Nil(t, err)
			_, err = semver.ParseRange(test.Version)
			assert.Nil(t, err)
			assert.Equal(t, len(test.OutputVersions), len(versions), test.Version)
			for _, ver := range test.OutputVersions {
				assert.Contains(t, versions, ver)
			}
		} else {
			assert.NotNil(t, err)
		}
	}
}
