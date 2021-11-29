package languages

type Config struct {
	Language string
	NameOverrides map[string]string
	Variables     []BuilderVar
}

type BuilderVar struct {
	Name string
	Description string
	VarType string
}