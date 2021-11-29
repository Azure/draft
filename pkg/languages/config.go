package languages

type Config struct {
	Language string
	NameOverrides []FileNameOverride
	Variables     []BuilderVar
}

type FileNameOverride struct {
	Path string
	Prefix string
}

type BuilderVar struct {
	Name string
	Description string
	VarType string
}