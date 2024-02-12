package v100

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"
)

var Params = map[string]*armpolicy.ParameterValuesValue{
	"allowedUsers": {
		Value: []string{"nodeclient", "system:serviceaccount:kube-system:aci-connector-linux", "system:serviceaccount:kube-system:node-controller", "acsService", "aksService", "system:serviceaccount:kube-system:cloud-node-manager"},
	},
	"allowedGroups": {
		Value: []string{"system:node"},
	},
	"cpuLimit": {
		Value: "200m",
	},
	"memoryLimit": {
		Value: "1Gi",
	},
	"excludedContainers": {
		Value: []string{},
	},
	"excludedImages": {
		Value: []string{},
	},
	"labels": {
		Value: []string{"kubernetes.azure.com"},
	},
	"allowedContainerImagesRegex": {
		Value: ".*",
	},
	"reservedTaints": {
		Value: []string{"CriticalAddonsOnly"},
	},
	"requiredProbes": {
		Value: []string{"readinessProbe", "livenessProbe"},
	},
}
