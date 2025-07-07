package example

import (
	"fmt"

	"github.com/Azure/draft/pkg/handlers"
	"github.com/Azure/draft/pkg/templatewriter/writers"
)

// WriteClusterResourcePlacementFilesExample shows how to set up a fileWriter and generate a fileMap using WriteClusterResourcePlacementFiles for Kubefleet
func WriteClusterResourcePlacementFilesExample() error {
	// Create a file map
	fileMap := make(map[string][]byte)

	// Create a template writer that writes to the file map
	w := writers.FileMapWriter{
		FileMap: fileMap,
	}

	// Select the kubefleet addon template type
	templateType := "kubefleet-clusterresourceplacement"

	// Create a map of inputs to the template (must correspond to the inputs in the template/addons/kubefleet/clusterresourceplacement/draft.yaml file)
	templateVars := map[string]string{
		"CRP_NAME":               "example-crp",
		"RESOURCE_SELECTOR_NAME": "example-namespace",
		"PLACEMENT_TYPE":         "PickFixed",
		"CLUSTER_NAMES":          "cluster-01,cluster-02,cluster-03",
		"PARTOF":                 "example-project",
		"GENERATORLABEL":         "draft",
	}

	// Set the output path for the ClusterResourcePlacement files
	outputPath := "./"

	// Get the kubefleet template
	template, err := handlers.GetTemplate(templateType, "", outputPath, &w)
	if err != nil {
		return fmt.Errorf("failed to get template: %e", err)
	}
	if template == nil {
		return fmt.Errorf("template is nil")
	}

	// Set the variable values within the template
	for k, v := range templateVars {
		template.Config.SetVariable(k, v)
	}

	// Generate the ClusterResourcePlacement files
	err = template.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate manifest: %e", err)
	}

	// Read written files from the file map
	fmt.Printf("Files written in WriteClusterResourcePlacementFilesExample:\n")
	for filePath, fileContents := range fileMap {
		if fileContents == nil {
			return fmt.Errorf("file contents for %s is nil", filePath)
		}
		fmt.Printf("  %s\n", filePath) // Print the file path
	}

	return nil
}