package example

import (
	"fmt"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/deployments"
	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
)

// WriteDeploymentFiles generates Deployment Files using Draft, writing to a Draft TemplateWriter. See the corresponding draft.yaml file in templates/deployments/[deployType] for the template inputs.
func WriteDeploymentFiles(w templatewriter.TemplateWriter, deploymentOutputPath string, deployConfig *config.DraftConfig, deploymentType string) error {
	d := deployments.CreateDeploymentsFromEmbedFS(template.Deployments, deploymentOutputPath)

	err := d.CopyDeploymentFiles(deploymentType, deployConfig, w)
	if err != nil {
		return fmt.Errorf("failed to generate manifest: %e", err)
	}
	return nil
}

// WriteDeploymentFilesExample shows how to set up a fileWriter and generate a fileMap using WriteDeploymentFiles
func WriteDeploymentFilesExample() error {
	// Create a file map
	fileMap := make(map[string][]byte)

	// Create a template writer that writes to the file map
	w := writers.FileMapWriter{
		FileMap: fileMap,
	}

	// Select the deployment type to generate the files for (must correspond to a directory in the template/deployments directory)
	deploymentType := "manifests"

	// Create a DraftConfig of inputs to the template (must correspond to the inputs in the template/deployments/<deploymentType>/draft.yaml file)
	deployConfig := &config.DraftConfig{
		Variables: []*config.BuilderVar{
			{
				Name:  "PORT",
				Value: "8080",
			},
			{
				Name:  "APPNAME",
				Value: "example-app",
			},
			{
				Name:  "SERVICEPORT",
				Value: "8080",
			},
			{
				Name:  "NAMESPACE",
				Value: "example-namespace",
			},
			{
				Name:  "IMAGENAME",
				Value: "example-image",
			},
			{
				Name:  "IMAGETAG",
				Value: "latest",
			},
		},
	}

	// Set the output path for the deployment files
	outputPath := "./"

	// Write the deployment files
	err := WriteDeploymentFiles(&w, outputPath, deployConfig, deploymentType)
	if err != nil {
		return err
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
