package example

import (
	"fmt"

	"github.com/Azure/draft/pkg/handlers"
	"github.com/Azure/draft/pkg/templatewriter/writers"
)

// WriteDeploymentFilesExample shows how to set up a fileWriter and generate a fileMap using WriteDeploymentFiles
func WriteDeploymentFilesExample() error {
	// Create a file map
	fileMap := make(map[string][]byte)

	// Create a template writer that writes to the file map
	w := writers.FileMapWriter{
		FileMap: fileMap,
	}

	// Select the deployment type to generate the files for (must correspond to a directory in the template/deployments directory)
	deploymentTemplateType := "deployment-manifests"

	// Create a map of of inputs to the template (must correspond to the inputs in the template/deployments/<deploymentType>/draft.yaml files)
	templateVars := map[string]string{
		"PORT":           "8080",
		"APPNAME":        "example-app",
		"SERVICEPORT":    "8080",
		"NAMESPACE":      "example-namespace",
		"IMAGENAME":      "example-image",
		"IMAGETAG":       "latest",
		"GENERATORLABEL": "draft",
	}

	// Set the output path for the deployment files
	outputPath := "./"

	// Get the deployment template
	d, err := handlers.GetTemplate(deploymentTemplateType, "", outputPath, &w)
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

	// Generate the deployment files
	err = d.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate manifest: %e", err)
	}

	// Read written files from the file map
	fmt.Printf("Files written in WriteDeploymentFilesExample:\n")
	for filePath, fileContents := range fileMap {
		if fileContents == nil {
			return fmt.Errorf("file contents for %s is nil", filePath)
		}
		fmt.Printf("  %s\n", filePath) // Print the file path
	}

	return nil
}
