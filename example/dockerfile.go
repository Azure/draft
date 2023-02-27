package example

import (
	"fmt"

	"github.com/Azure/draft/pkg/languages"
	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/template"
)

// WriteDockerfile generates a Dockerfile using Draft, writing to a Draft TemplateWriter. See the corresponding draft.yaml file for the template inputs.
func WriteDockerfile(w templatewriter.TemplateWriter, dockerfileOutputPath string, dockerfileInputs map[string]string, generationLanguage string) error {
	l := languages.CreateLanguagesFromEmbedFS(template.Dockerfiles, dockerfileOutputPath)

	err := l.CreateDockerfileForLanguage(generationLanguage, dockerfileInputs, w)
	if err != nil {
		return fmt.Errorf("failed to generate dockerfile: %e", err)
	}
	return nil
}
