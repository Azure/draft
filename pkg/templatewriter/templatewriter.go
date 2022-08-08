package templatewriter

type TemplateWriter interface {
	WriteFile(string, []byte) error
	EnsureDirectory(string) error
}
