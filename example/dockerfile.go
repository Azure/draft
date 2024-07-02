package example

import (
	"fmt"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/languages"
	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
)

// WriteDockerfile generates a Dockerfile and dockerignore using Draft, writing to a Draft TemplateWriter. See the corresponding draft.yaml file in templates/dockerfiles/[language] for the template inputs.
func WriteDockerfile(w templatewriter.TemplateWriter, dockerfileOutputPath string, langConfig *config.DraftConfig, generationLanguage string) error {
	l := languages.CreateLanguagesFromEmbedFS(template.Dockerfiles, dockerfileOutputPath)

	err := l.CreateDockerfileForLanguage(generationLanguage, langConfig, w)
	if err != nil {
		return fmt.Errorf("failed to generate dockerfile: %e", err)
	}
	return nil
}

// WriteDockerfileExample shows how to set up a fileWriter and generate a fileMap using WriteDockerfile
func WriteDockerfileExample() error {
	// Create a file map
	fileMap := make(map[string][]byte)

	// Create a template writer that writes to the file map
	w := writers.FileMapWriter{
		FileMap: fileMap,
	}

	// Select the language to generate the Dockerfile for (must correspond to a directory in the template/dockerfiles directory)
	generationLanguage := "go"

	// Create a DraftConfig of inputs to the template (must correspond to the inputs in the template/dockerfiles/<language>/draft.yaml file)
	langConfig := &config.DraftConfig{
		Variables: []*config.BuilderVar{
			{
				Name:  "PORT",
				Value: "8080",
			},
			{
				Name:  "VERSION",
				Value: "1.20",
			},
		},
	}

	// Set the output path for the Dockerfile
	outputPath := "./"

	// Write the Dockerfile
	err := WriteDockerfile(&w, outputPath, langConfig, generationLanguage)
	if err != nil {
		return err
	}

	// Read written files from the file map
	fmt.Printf("Files written in WriteDockerfileExample:\n")
	for filePath, fileContents := range fileMap {
		if fileContents == nil {
			return fmt.Errorf("file contents for %s is nil", filePath)
		}
		fmt.Printf("  %s\n", filePath) // Print the file path
	}

	return nil
}
