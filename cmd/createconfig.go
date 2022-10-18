package cmd

type CreateConfig struct {
	DeployType        string
	LanguageType      string
	DeployVariables   []UserInputs
	LanguageVariables []UserInputs
}

type UserInputs struct {
	Name  string
	Value string
}
