package defaults

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/Azure/draft/pkg/reporeader"
)

func TestGradleExtractor_ReadDefaults(t *testing.T) {
	content, err := ioutil.ReadFile("testdata/sample.gradle")
	if err != nil {
		t.Errorf("error reading sample_build.gradle: %v", err)
	}
	type args struct {
		r reporeader.RepoReader
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "extract gradle jre version with spaces",
			args: args{
				r: reporeader.TestRepoReader{
					Files: map[string][]byte{
						"build.gradle": []byte("group = 'com.example' version = '0.0.1-SNAPSHOT' sourceCompatibility=\"11\" targetCompatibility='11'"),
					},
				},
			},
			want: map[string]string{
				"VERSION":        "11-jre",
				"BUILDERVERSION": "jdk11",
			},
			wantErr: false,
		},
		{
			name: "extract gradle jre version with new lines",
			args: args{
				r: reporeader.TestRepoReader{
					Files: map[string][]byte{
						"build.gradle": []byte("group = 'com.example'\nversion = '0.0.1-SNAPSHOT'\nsourceCompatibility=\"11\"\ntargetCompatibility='11'"),
					},
				},
			},
			want: map[string]string{
				"VERSION":        "11-jre",
				"BUILDERVERSION": "jdk11",
			},
			wantErr: false,
		},
		{
			name: "extract gradle jre version with tabs",
			args: args{
				r: reporeader.TestRepoReader{
					Files: map[string][]byte{
						"build.gradle": []byte("group = 'com.example'\tversion = '0.0.1-SNAPSHOT'\tsourceCompatibility= \"12\" \ntargetCompatibility='11'"),
					},
				},
			},
			want: map[string]string{
				"VERSION":        "12-jre",
				"BUILDERVERSION": "jdk11",
			},
			wantErr: false,
		},
		{
			name: "extract gradle jre version with double spaces",
			args: args{
				r: reporeader.TestRepoReader{
					Files: map[string][]byte{
						"build.gradle": []byte("group = 'com.example'  version = '0.0.1-SNAPSHOT'  sourceCompatibility=\"12\"\ntargetCompatibility='11'"),
					},
				},
			},
			want: map[string]string{
				"VERSION":        "12-jre",
				"BUILDERVERSION": "jdk11",
			},
			wantErr: false,
		},
		{
			name: "extract gradle jre version reading from a file",
			args: args{
				r: reporeader.TestRepoReader{
					Files: map[string][]byte{
						"build.gradle": content,
					},
				},
			},
			want: map[string]string{
				"VERSION":        "11-jre",
				"BUILDERVERSION": "jdk11",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := GradleExtractor{}
			got, err := p.ReadDefaults(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadDefaults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadDefaults() got = %v, want %v", got, tt.want)
			}
		})
	}

}
