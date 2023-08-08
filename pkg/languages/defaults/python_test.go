package defaults

import (
	"reflect"
	"testing"

	"github.com/Azure/draft/pkg/reporeader"
)

func TestPythonExtractor_MatchesLanguage(t *testing.T) {
	type args struct {
		lowerlang string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "lowercase python",
			args: args{
				lowerlang: "python",
			},
			want: true,
		},
		{
			name: "shouldn't match go",
			args: args{
				lowerlang: "go",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PythonExtractor{}
			if got := p.MatchesLanguage(tt.args.lowerlang); got != tt.want {
				t.Errorf("MatchesLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPythonExtractor_ReadDefaults(t *testing.T) {
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
			name: "extract first python file as entrypoint",
			args: args{
				r: reporeader.FakeRepoReader{
					Files: map[string][]byte{
						"foo.py": []byte("print('hello world')"),
						"bar.py": []byte("print('hello world')"),
					},
				},
			},
			want: map[string]string{
				"ENTRYPOINT": "foo.py",
			},
			wantErr: false,
		},
		{
			name: "extract python file containing the string \"if __name__ == '__main__'\" as the entrypoint",
			args: args{
				r: reporeader.FakeRepoReader{
					Files: map[string][]byte{
						"foo.py": []byte("print('hello world')"),
						"bar.py": []byte("if __name__ == '__main__' : print('hello world')"),
					},
				},
			},
			want: map[string]string{
				"ENTRYPOINT": "bar.py",
			},
			wantErr: false,
		},
		{
			name: "extract python file containing the string \"if __name__==\"__main__\"\" as the entrypoint",
			args: args{
				r: reporeader.FakeRepoReader{
					Files: map[string][]byte{
						"foo.py": []byte("print('hello world')"),
						"bar.py": []byte("if __name__==\"__main__\": print('hello world')"),
					},
				},
			},
			want: map[string]string{
				"ENTRYPOINT": "bar.py",
			},
			wantErr: false,
		},
		{
			name: "extract python file named app.py as the entrypoint",
			args: args{
				r: reporeader.FakeRepoReader{
					Files: map[string][]byte{
						"foo.py": []byte("print('Hello World')"),
						"app.py": []byte("print('Hello World')"),
					},
				},
			},
			want: map[string]string{
				"ENTRYPOINT": "app.py",
			},
			wantErr: false,
		},
		{
			name: "no extraction if no python files",
			args: args{
				r: reporeader.FakeRepoReader{
					Files: map[string][]byte{
						"foo.notpy": []byte("print('hello world')"),
						"bar":       []byte("print('hello world')"),
					},
				},
			},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "empty extraction with no files",
			args: args{
				r: reporeader.FakeRepoReader{
					Files: map[string][]byte{},
				},
			},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "ignore files below depth root depth",
			args: args{
				r: reporeader.FakeRepoReader{
					Files: map[string][]byte{
						"dir/foo.py": []byte("print('hello world')"),
					},
				},
			},
			want:    map[string]string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PythonExtractor{}
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
