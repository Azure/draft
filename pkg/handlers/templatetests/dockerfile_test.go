package templatetests

import (
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestDockerfileTemplates(t *testing.T) {
	tests := []TestInput{
		{
			Name:            "valid clojure dockerfile",
			TemplateName:    "dockerfile-clojure",
			FixturesBaseDir: "../../fixtures/dockerfiles/clojure",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":    "80",
				"VERSION": "19-jdk-alpine",
			},
		},
		{
			Name:            "valid csharp dockerfile",
			TemplateName:    "dockerfile-csharp",
			FixturesBaseDir: "../../fixtures/dockerfiles/csharp",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":    "80",
				"VERSION": "6.0",
			},
		},
		{
			Name:            "valid erlang dockerfile",
			TemplateName:    "dockerfile-erlang",
			FixturesBaseDir: "../../fixtures/dockerfiles/erlang",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":         "80",
				"BUILDVERSION": "27.0-alpine",
				"VERSION":      "3.17",
			},
		},
		{
			Name:            "valid go dockerfile",
			TemplateName:    "dockerfile-go",
			FixturesBaseDir: "../../fixtures/dockerfiles/go",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":    "80",
				"VERSION": "1.23",
			},
		},
		{
			Name:            "valid gomodule dockerfile",
			TemplateName:    "dockerfile-gomodule",
			FixturesBaseDir: "../../fixtures/dockerfiles/gomodule",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":    "80",
				"VERSION": "1.23",
			},
		},
		{
			Name:            "valid gradle dockerfile",
			TemplateName:    "dockerfile-gradle",
			FixturesBaseDir: "../../fixtures/dockerfiles/gradle",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":         "80",
				"BUILDVERSION": "jdk21",
				"VERSION":      "21-jre",
			},
		},
		{
			Name:            "valid gradlew dockerfile",
			TemplateName:    "dockerfile-gradlew",
			FixturesBaseDir: "../../fixtures/dockerfiles/gradlew",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":         "80",
				"BUILDVERSION": "jdk21",
				"VERSION":      "21-jre",
			},
		},
		{
			Name:            "valid java dockerfile",
			TemplateName:    "dockerfile-java",
			FixturesBaseDir: "../../fixtures/dockerfiles/java",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":         "80",
				"BUILDVERSION": "3 (jdk-21)",
				"VERSION":      "21-jre",
			},
		},
		{
			Name:            "valid javascript dockerfile",
			TemplateName:    "dockerfile-javascript",
			FixturesBaseDir: "../../fixtures/dockerfiles/javascript",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":    "80",
				"VERSION": "14.15.4",
			},
		},
		{
			Name:            "valid php dockerfile",
			TemplateName:    "dockerfile-php",
			FixturesBaseDir: "../../fixtures/dockerfiles/php",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":         "80",
				"BUILDVERSION": "1",
				"VERSION":      "7.1-apache",
			},
		},
		{
			Name:            "valid python dockerfile",
			TemplateName:    "dockerfile-python",
			FixturesBaseDir: "../../fixtures/dockerfiles/python",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":       "80",
				"ENTRYPOINT": "app.py",
				"VERSION":    "3.9",
			},
		},
		{
			Name:            "valid ruby dockerfile",
			TemplateName:    "dockerfile-ruby",
			FixturesBaseDir: "../../fixtures/dockerfiles/ruby",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":    "80",
				"VERSION": "3.1.2",
			},
		},
		{
			Name:            "valid rust dockerfile",
			TemplateName:    "dockerfile-rust",
			FixturesBaseDir: "../../fixtures/dockerfiles/rust",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":    "80",
				"VERSION": "1.70.0",
			},
		},
		{
			Name:            "valid swift dockerfile",
			TemplateName:    "dockerfile-swift",
			FixturesBaseDir: "../../fixtures/dockerfiles/swift",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"PORT":    "80",
				"VERSION": "5.5",
			},
		},
	}

	for _, test := range tests {
		RunTemplateTest(t, test)
	}
}
