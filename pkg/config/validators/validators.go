package validators

func GetValidator(variableKind string) func(string) error {
	switch variableKind {
	default:
		return DefaultValidator
	}
}

func DefaultValidator(input string) error {
	return nil
}
