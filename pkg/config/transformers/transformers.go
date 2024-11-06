package transformers

import (
	"encoding/json"
	"fmt"
)

func GetTransformer(variableKind string) func(string) (any, error) {
	switch variableKind {
	case "envVarMap":
		return EnvironmentVariableMapTransformer
	default:
		return DefaultTransformer
	}
}

func EnvironmentVariableMapTransformer(inputVar string) (any, error) {
	var inputVarMap map[string]string
	if err := json.Unmarshal([]byte(inputVar), &inputVarMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal variable as map[string]string: %s", err)
	}
	return inputVarMap, nil
}

func DefaultTransformer(inputVar string) (any, error) {
	return inputVar, nil
}
