package dryrun

type DryRunInfo struct {
	Variables    map[string]string `json:"variables"`
	FilesToWrite []string          `json:"filesToWrite"`
}

type DryRunWriter struct {
	DryRunInfo *DryRunInfo
}

func (d *DryRunWriter) WriteFile(path string, data []byte) error {
	d.DryRunInfo.FilesToWrite = append(d.DryRunInfo.FilesToWrite, path)
	return nil
}

func (d *DryRunWriter) EnsureDirectory(path string) error {
	return nil
}

func (d *DryRunWriter) SetVariable(key, value string) {
	d.DryRunInfo.Variables[key] = value
}

func NewDryRunWriter() *DryRunWriter {
	return &DryRunWriter{
		DryRunInfo: &DryRunInfo{
			Variables:    make(map[string]string),
			FilesToWrite: make([]string, 0),
		},
	}
}
