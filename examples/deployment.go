package examples

import (
	"fmt"

	"github.com/Azure/draft/pkg/deployments"
	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/template"
)

// WriteDeploymentFiles generates Deployment Files using Draft, writing to a Draft TemplateWriter. See the corresponding draft.yaml file for the template inputs.
func WriteDeploymentFiles(deploymentOutputPath string, w templatewriter.TemplateWriter, deploymentInputs map[string]string, deploymentType string) error {
	d := deployments.CreateDeploymentsFromEmbedFS(template.Deployments, deploymentOutputPath)

	err := d.CopyDeploymentFiles(deploymentType, deploymentInputs, w)
	if err != nil {
		return fmt.Errorf("failed to generate manifest: %e", err)
	}
	return nil
}
