package example

import (
	"fmt"

	"github.com/Azure/draft/pkg/handlers"
	"github.com/Azure/draft/pkg/templatewriter/writers"
)

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

	// Create a map of inputs to the template (must correspond to the inputs in the template/dockerfiles/<language>/draft.yaml files)
	templateVars := map[string]string{
		"PORT":           "8080",
		"APPNAME":        "example-app",
		"SERVICEPORT":    "8080",
		"NAMESPACE":      "example-namespace",
		"IMAGENAME":      "example-image",
		"IMAGETAG":       "latest",
		"GENERATORLABEL": "draft",
	}

	// Set the output path for the Dockerfile
	outputPath := "./"

	// Get the dockerfile template
	d, err := handlers.GetTemplate(fmt.Sprintf("dockerfile-%s", generationLanguage), "", outputPath, &w)
	if err != nil {
		return fmt.Errorf("failed to get template: %e", err)
	}
	if d == nil {
		return fmt.Errorf("template is nil")
	}

	// Set the variable values within the template
	for k, v := range templateVars {
		d.Config.SetVariable(k, v)
	}

	// Generate the dockerfile files
	err = d.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate dockerfile: %e", err)
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
