package safeguards

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"os"
	"path"
	"strings"

	"github.com/open-policy-agent/opa/rego"
	"gopkg.in/yaml.v3"
)

// Constants
const (
	Constraint_CAI = "container-allowed-images"
	Constraint_CEP = "container-enforce-probes"
	Constraint_CRL = "container-resource-limits"
	Constraint_NUP = "no-unauthenticated-pulls"
	Constraint_PDB = "pod-disruption-budgets"
	Constraint_PEA = "pod-enforce-antiaffinity"
	Constraint_RT  = "restricted-taints"
	Constraint_USS = "unique-service-selectors"
)

type Safeguard struct {
	name     string
	filepath string
}

var supportedSafeguards = []Safeguard{
	{
		name:     Constraint_CAI,
		filepath: "constraints/ContainerAllowedImages/container-allowed-images.yaml",
	},
	{
		name:     Constraint_CEP,
		filepath: "constraints/ContainerEnforceProbes/container-allowed-images.yaml",
	},
	{
		name:     Constraint_CRL,
		filepath: "constraints/ContainerResourceLimits/container-resource-limits.yaml",
	},
	{
		name:     Constraint_NUP,
		filepath: "constraints/NoUnauthenticatedPulls/no-unauthenticated-pulls.yaml",
	},
	{
		name:     Constraint_PDB,
		filepath: "constraints/PodDisruptionBudgets/pod-disruption-budgets.yaml",
	},
	{
		name:     Constraint_PEA,
		filepath: "constraints/PodEnforceAntiaffinity/pod-enforce-antiaffinity.yaml",
	},
	{
		name:     Constraint_RT,
		filepath: "constraints/RestrictedTaints/restricted-taints.yaml",
	},
	{
		name:     Constraint_USS,
		filepath: "constraints/UniqueServiceSelectors/unique-service-selectors.yaml",
	},
}

// ConstraintFetcher is the interface used to fetch each safeguards constraint
type ConstraintFetcher interface {
	Fetch() ([]ConstraintFile, map[string]appsv1.Deployment, error)
}

type DeploymentFile struct {
	Metadata DeploymentMetadata `yaml:"metadata"`
	Spec     DeploymentSpec     `yaml:"spec"`
}
type DeploymentMetadata struct {
	Name string `yaml:"name"`
	Labels
}
type MetadataLabels struct{}
type DeploymentSpec struct {
	Replicas string   `yaml:"replicas"`
	Selector Selector `yaml:"selector"`
	Template Template `yaml:"template"`
}
type Selector struct {
	MatchLabels MatchLabels `yaml:"matchLabels"`
}
type MatchLabels struct {
	App string `yaml:"app"`
}
type Template struct {
	Metadata TemplateMetadata `yaml:"metadata"`
	Spec     TemplateSpec     `yaml:"spec"`
}
type TemplateMetadata struct {
	Labels Labels `yaml:"labels"`
}
type Labels struct {
	App string `yaml:"app"`
}
type TemplateSpec struct {
	InitContainers []InitContainers `yaml:"initContainers"`
	Containers     []Containers     `yaml:"containers"`
}
type InitContainers struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}
type Containers struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}

// ConstraintFile is our struct implementation of the safeguards constraint YAML
// create a getParameters() method
type ConstraintFile struct {
	Metadata Metadata `yaml:"metadata"`
	Spec     Spec     `yaml:"spec"`
	Name     string
}
type Spec struct {
	Targets []Target `yaml:"targets"`
}
type Target struct {
	Target string   `yaml:"target"`
	Rego   string   `yaml:"rego"`
	Libs   []string `yaml:"libs"`
}
type Metadata struct {
	Name string `yaml:"name"`
}

// ConstraintsBuilderA is the implementation of ConstraintFetcher that reads in constraints from the local fs
type FilesystemConstraintFetcher struct {
}

// fetchTestDeploymentFile pulls in our test deployment YAML
func fetchDeploymentFile(deploymentPath string) (map[string]interface{}, error) {
	wd, _ := os.Getwd()
	completePath := path.Join(wd, deploymentPath)
	bs, err := os.ReadFile(completePath)
	if err != nil {
		// handle error
		return nil, fmt.Errorf("reading deployment: %w", err)
	}

	var deploymentFile map[string]interface{}
	if err := yaml.Unmarshal(bs, &deploymentFile); err != nil {
		// handle error
		return nil, fmt.Errorf("unmarshaling input: %w", err)
	}

	return deploymentFile, nil
}

type Params map[string]interface{}

func buildParams(constraint string) Params {

	//var excludedContainers []string

	// thbarnes: this is where we manipulate params to properly test rego validation
	if constraint == Constraint_CAI {
		//excludedContainers = []string{"my-container-CAI-error"}
	} else if constraint == Constraint_CEP {

	} else if constraint == Constraint_CRL {

	} else if constraint == Constraint_NUP {

	} else if constraint == Constraint_PDB {

	} else if constraint == Constraint_PEA {

	} else if constraint == Constraint_RT {

	} else if constraint == Constraint_USS {

	}

	p := Params{
		"allowedUsers": []string{
			"nodeclient",
			"system:serviceaccount:kube-system:aci-connector-linux",
			"system:serviceaccount:kube-system:node-controller",
			"acsService",
			"aksService",
			"system:serviceaccount:kube-system:cloud-node-manager",
		},
		"allowedGroups": []string{
			"system:node",
		},
		"cpuLimit":           "200m",
		"memoryLimit":        "1Gi",
		"excludedContainers": []string{"my-container-CAI-error"},
		"excludedImages":     []string{},
		"labels": []string{
			"kubernetes.azure.com",
		},
		"allowedContainerImagesRegex": ".*",
		"reservedTaints": []string{
			"CriticalAddonsOnly",
		},
		"requiredProbes": []string{
			"readinessProbe",
			"livenessProbe",
		},
		"imageRegex": "<something>",
	}

	return p
}

// thbarnes: placeholder for now
func buildUserInfo() map[string]interface{} {
	u := map[string]interface{}{
		"username": "system:serviceaccount:kube-system:replicaset-controller",
		"uid":      "439dea65-3e4e-4fa8-b5f8-8fdc4bc7cf53",
		"groups": []string{
			"system:serviceaccounts",
			"system:serviceaccounts:kube-system",
			"system:authenticated",
		},
	}

	return u
}

// buildInput creates our input JSON when given a deployment file
func buildInput(deployment map[string]interface{}, constraint string) map[string]interface{} {
	// thbarnes: buildInput this needs to be refined and automated
	// generated by kuberenetes api server -> not necessary for testing
	// would be good to make the input a struct

	input := map[string]interface{}{
		"review": map[string]interface{}{
			"object":   deployment,
			"userInfo": buildUserInfo(),
		},
		"parameters": buildParams(constraint),
	}
	return input
}

// Option A's method of fetching constraints
// thbarnes: refine the path code
func (fcf FilesystemConstraintFetcher) Fetch() ([]ConstraintFile, error) {
	// list of constraint files to be read in and queried
	var c []ConstraintFile

	wd, _ := os.Getwd()
	for _, s := range supportedSafeguards {
		completePath := path.Join(wd, s.filepath)
		b, err := os.ReadFile(completePath)
		if err != nil {
			return c, fmt.Errorf("reading constraint file: %s", s.name)
		}

		var constraintFile ConstraintFile
		if err := yaml.Unmarshal(b, &constraintFile); err != nil {
			return c, fmt.Errorf("unmarshaling constraint: %w", err)
		}

		c = append(c, constraintFile)
	}

	return c, nil
}

func getConstraintFileName(path string) string {
	splitPath := strings.Split(path, "/")
	return strings.Split(splitPath[len(splitPath)], ".yaml")[0]
}

func (fcf FilesystemConstraintFetcher) FetchOne(name string) (ConstraintFile, error) {
	// list of constraint files to be read in and queried
	var c ConstraintFile

	for _, s := range supportedSafeguards {
		if s.name == name {
			wd, _ := os.Getwd()
			completePath := path.Join(wd, s.filepath)
			b, err := os.ReadFile(completePath)
			if err != nil {
				return c, fmt.Errorf("reading constraint file: %s", s.name)
			}

			var constraintFile ConstraintFile
			if err := yaml.Unmarshal(b, &constraintFile); err != nil {
				return c, fmt.Errorf("unmarshaling constraint: %w", err)
			}

			constraintFile.Name = getConstraintFileName(completePath)
			c = constraintFile
		}
	}

	return c, nil
}

// sanitizeRegoPolicy removes problematic lines from our rego code for consumption as our rego policy in the evalution step
func sanitizeRegoPolicy(rp string) string {
	lines := strings.Split(rp, "\n")
	var newLines []string
	for _, line := range lines {
		if !strings.Contains(line, "import data.lib.") && !strings.Contains(line, "package lib.") {
			newLines = append(newLines, line)
		}
	}
	return strings.Join(newLines, "\n")
}

// appendLibs appends every lib item from the constraint YAML, separated by newlines
func appendLibs(libs []string) string {
	l := ""
	if len(libs) > 0 {
		for _, lib := range libs {
			l += lib + "\n"
		}
	}

	return l
}

func buildQueryString(name string) string {
	return "x = data." + name + ".violation"
}

func evaluateQuery(ctx context.Context, file ConstraintFile, deployment map[string]interface{}) error {
	queryString := buildQueryString(file.Metadata.Name)

	// thbarnes: throw in a check for if length is 0 or >1 and error if so
	l := appendLibs(file.Spec.Targets[0].Libs)
	regoString := file.Spec.Targets[0].Rego
	regoPolicy := sanitizeRegoPolicy(regoString + l)

	r := rego.New(
		rego.Query(queryString),
		rego.Module("main.rego", regoPolicy))

	query, err := r.PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("creating rego query: %w", err)
	}

	// our input state
	// build inputs PER CONSTRAINT

	input := buildInput(deployment, file.Name)

	// thbarnes: investigate if you can make your own `input` struct and pass it into here
	rs, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("evaluating query: %w", err)
	}

	fmt.Println("Result:", rs[0].Bindings["x"])

	return nil
}

// ValidateDeployment is what will be called by `draft validate` to validate the user's deployment manifest
// against each safeguards constraint
func ValidateDeployment(deploymentPath, constraint string) error {
	// thbarnes: ConstraintsBuilderB will eventually take over
	ctx := context.Background()

	var fcf FilesystemConstraintFetcher

	constraintFiles, err := fcf.Fetch()
	if err != nil {
		return fmt.Errorf("fetching constraints: %w", err)
	}

	deployment, err := fetchDeploymentFile(deploymentPath)

	for _, file := range constraintFiles {
		err = evaluateQuery(ctx, file, deployment)
		if err != nil {
			return fmt.Errorf("evaluating query: %w", err)
		}
	}

	return nil
}
