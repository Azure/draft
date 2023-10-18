package guardrails

import (
	"context"
	"fmt"
	"os"

	"github.com/open-policy-agent/opa/rego"
	"gopkg.in/yaml.v3"
)

// Constants
const ConstraintsDirectory = "./constraints"
const DeploymentFileDirectory = "./deployment/deployment.yaml"

// ConstraintFetcher is the interface used to fetch each guardrails constraint
type ConstraintFetcher interface {
	Fetch() (*[]ConstraintFile, error)
}

// ConstraintFile is our struct implementation of the guardrails constraint YAML
type ConstraintFile struct {
	Metadata Metadata `yaml:"metadata"`
	Targets  Targets  `yaml:"targets"`
}
type Targets struct {
	Rego struct {
		Content string
	} `yaml:"rego"`
}
type Metadata struct {
	Name string `yaml:"name"`
}

// ConstraintsBuilderA is the implementation of ConstraintFetcher that reads in constraints from the local fs
type ConstraintsBuilderA struct {
}

// fetchDemoDeploymentFile pulls in our example deployment YAML
func fetchDemoDeploymentFile() map[string]interface{} {
	bs, err := os.ReadFile(DeploymentFileDirectory) // thbarnes: need to refine where we get this, for now hard-code
	if err != nil {
		// handle error
		fmt.Errorf("reading deployment:", err.Error())
	}

	var deploymentFile map[string]interface{}
	if err := yaml.Unmarshal(bs, &deploymentFile); err != nil {
		// handle error
		fmt.Errorf("unmarshaling input:", err.Error())
	}

	return deploymentFile
}

// buildInput creates our input JSON when given a deployment file
func buildInput(df map[string]interface{}) map[string]interface{} {
	// Define input data (your Deployment YAML).

	// thbarnes: this needs to be refined and automated
	input := map[string]interface{}{
		"review": map[string]interface{}{
			"object": df,
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
func (cba ConstraintsBuilderA) Fetch() (*[]ConstraintFile, error) {
	var c *[]ConstraintFile

	constraints, err := os.ReadDir(ConstraintsDirectory)
	if err != nil {
		return nil, fmt.Errorf("reading guardrails constraints directory")
	}

	for _, con := range constraints {
		fullConstraintDir := ConstraintsDirectory + con.Name()
		b, err := os.ReadFile(fullConstraintDir)
		if err != nil {
			return nil, fmt.Errorf("reading constraint file:" + con.Name())
		}

		var constraintFile ConstraintFile
		if err := yaml.Unmarshal(b, &constraintFile); err != nil {
			fmt.Errorf("unmarshaling constraint:", err.Error())
		}

		*c = append(*c, constraintFile)
	}

	return c, nil
}

// validateGuardrailsConstraints is what will be called by `draft validate` to validate the user's deployment manifest
// against each guardrails constraint
func validateGuardrailsConstraint(ctx context.Context) {
	// thbarnes: ConstraintsBuilderB will eventually take over
	var cf ConstraintFetcher
	cb := ConstraintsBuilderA{}
	cf = cb

	constraintFiles, err := cf.Fetch()
	if err != nil {
		fmt.Errorf("fetching constraints")
	}

	// our input state
	input := buildInput(fetchDemoDeploymentFile())

	// evaluate each rego policy against the deployment file
	for _, policy := range *constraintFiles {
		queryString := "x = data." + policy.Metadata.Name + ".violation" // thbarnes: need a better way to qualify the rego func
		r := rego.New(
			rego.Query(queryString),
			rego.Module("main.rego", policy))

		query, err := r.PrepareForEval(ctx)
		if err != nil {
			fmt.Errorf("creating rego query:", err.Error())
		}

		rs, err := query.Eval(ctx, rego.EvalInput(input))
		if err != nil {
			// handle error
			fmt.Errorf("evaluating query:", err.Error())
		}

		fmt.Println("Result:", rs[0].Bindings["x"])
	}
}
