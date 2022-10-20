package cmd

type CreateConfig struct {
	DeployType        string       `yaml:"deployType"`
	LanguageType      string       `yaml:"languageType"`
	DeployVariables   []UserInputs `yaml:"deployVariables"`
	LanguageVariables []UserInputs `yaml:"languageVariables"`
}

type UserInputs struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
