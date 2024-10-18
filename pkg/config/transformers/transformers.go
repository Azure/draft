package transformers

func GetTransformer(variableKind string) func(string) (string, error) {
	switch variableKind {
	default:
		return DefaultTransformer
	}
}

func DefaultTransformer(inputVar string) (string, error) {
	return inputVar, nil
}
