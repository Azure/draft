package validators

import (
	"encoding/json"
	"fmt"
)

func GetValidator(variableKind string) func(string) error {
	switch variableKind {
	case "envVarMap":
		return keyValueMapValidator
	case "imagePullPolicy":
		return imagePullPolicyValidator
	case "kubernetesProbeType":
		return kubernetesProbeTypeValidator
	case "scalingResourceType":
		return scalingResourceTypeValidator
	default:
		return defaultValidator
	}
}

func imagePullPolicyValidator(input string) error {
	switch input {
	case "Always", "IfNotPresent", "Never":
		return nil
	default:
		return fmt.Errorf("invalid image pull policy: %s. valid values: Always, IfNotPresent, Never", input)
	}
}

func scalingResourceTypeValidator(input string) error {
	switch input {
	case "cpu", "memory":
		return nil
	default:
		return fmt.Errorf("invalid scaling resource type: %s. valid values: cpu, memory", input)
	}
}

func kubernetesProbeTypeValidator(input string) error {
	switch input {
	case "httpGet", "tcpSocket":
		return nil
	default:
		return fmt.Errorf("invalid probe type: %s. valid values: httpGet, tcpSocket", input)
	}
}

func keyValueMapValidator(input string) error {
	if err := json.Unmarshal([]byte(input), &map[string]string{}); err != nil {
		return fmt.Errorf("failed to unmarshal variable as map[string]string: %s", err)
	}
	return nil
}

func defaultValidator(input string) error {
	return nil
}
