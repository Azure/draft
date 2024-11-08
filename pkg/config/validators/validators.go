package validators

import (
	"encoding/json"
	"fmt"
)

func GetValidator(variableKind string) func(string) error {
	switch variableKind {
	case "envVarMap":
		return KeyValueMapValidator
	default:
		return DefaultValidator
	}
}

func KeyValueMapValidator(input string) error {
	if err := json.Unmarshal([]byte(input), &map[string]string{}); err != nil {
		return fmt.Errorf("failed to unmarshal variable as map[string]string: %s", err)
	}
	return nil
}

func DefaultValidator(input string) error {
	return nil
}
