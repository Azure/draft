package guardrails

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/open-policy-agent/opa/rego"
	"gopkg.in/yaml.v3"
)

// Constants
const ConstraintsDirectory = "./constraints/"
const DeploymentFilePrefix = "./deployment/"

// ConstraintFetcher is the interface used to fetch each guardrails constraint
type ConstraintFetcher interface {
	Fetch() ([]ConstraintFile, error)
}

// ConstraintFile is our struct implementation of the guardrails constraint YAML
type ConstraintFile struct {
	Metadata Metadata `yaml:"metadata"`
	Spec     Spec     `yaml:"spec"`
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
type ConstraintsBuilderA struct {
}

// fetchTestDeploymentFile pulls in our example deployment YAML
func fetchTestDeploymentFile(df string) (map[string]interface{}, error) {
	bs, err := os.ReadFile(DeploymentFilePrefix + df) // thbarnes: need to refine where we get this, for now hard-code
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

// buildInput creates our input JSON when given a deployment file
func buildInput(dfMap map[string]interface{}) map[string]interface{} {
	// thbarnes: this needs to be refined and automated
	input := map[string]interface{}{
		"review": map[string]interface{}{
			"object": dfMap,
			"userInfo": map[string]interface{}{
				"username": "system:serviceaccount:kube-system:replicaset-controller",
				"uid":      "439dea65-3e4e-4fa8-b5f8-8fdc4bc7cf53",
				"groups": []string{
					"system:serviceaccounts",
					"system:serviceaccounts:kube-system",
					"system:authenticated",
				},
			},
		},
		"parameters": map[string]interface{}{
			"allowedGroups": []string{
				"testGroup1",
				"testGroup2",
			},
			"allowedUsers": []string{
				"testUser1",
				"testUser2",
			},
			"labels": []string{
				"testLabel1",
				"testLabel3",
			},
		},
	}
	return input
}

// Option A's method of fetching constraints
// thbarnes: figure out why consts aren't being read into debug instance
// refine the path code
func (cba ConstraintsBuilderA) Fetch() ([]ConstraintFile, error) {
	var c []ConstraintFile
	cwd, _ := os.Getwd()
	dirs := []string{cwd, ConstraintsDirectory}
	fullPath := path.Join(dirs[0], dirs[1]) + "/"
	constraints, err := os.ReadDir(fullPath)
	if err != nil {
		return c, fmt.Errorf("reading guardrails constraints directory")
	}

	for _, con := range constraints {
		fullConstraintDir := path.Join(ConstraintsDirectory, con.Name())
		b, err := os.ReadFile(fullConstraintDir)
		if err != nil {
			return c, fmt.Errorf("reading constraint file: %s", con.Name())
		}

		var constraintFile ConstraintFile
		if err := yaml.Unmarshal(b, &constraintFile); err != nil {
			return c, fmt.Errorf("unmarshaling constraint: %w", err)
		}

		c = append(c, constraintFile)
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

// validateGuardrailsConstraints is what will be called by `draft validate` to validate the user's deployment manifest
// against each guardrails constraint
func ValidateGuardrailsConstraint(df string) error {
	// thbarnes: ConstraintsBuilderB will eventually take over
	ctx := context.TODO()

	var cf ConstraintFetcher
	cb := ConstraintsBuilderA{}
	cf = cb

	constraintFiles, err := cf.Fetch()
	if err != nil {
		return fmt.Errorf("fetching constraints: %w", err)
	}

	dfMap, err := fetchTestDeploymentFile(df)
	if err != nil {
		return fmt.Errorf("fetching test deployment: %w", err)
	}

	// our input state
	input := buildInput(dfMap)

	// evaluate each rego policy against the deployment file
	//thbarnes:
	// david suggested worker pattern to break out into goroutines to parallelize and aggregate the errors
	for _, file := range constraintFiles {
		queryString := "x = data." + file.Metadata.Name + ".violation" // thbarnes: need a better way to qualify the rego func
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

		rs, err := query.Eval(ctx, rego.EvalInput(input))
		if err != nil {
			return fmt.Errorf("evaluating query: %w", err)
		}

		fmt.Println("Result:", rs[0].Bindings["x"])
	}

	return nil
}
