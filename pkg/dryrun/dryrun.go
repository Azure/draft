package dryrun

type DryRunInfo struct {
	Variables    map[string]string `json:"variables"`
	FilesToWrite []string          `json:"filesToWrite"`
}

type DryRunRecorder struct {
	DryRunInfo *DryRunInfo
}

func (d *DryRunRecorder) WriteFile(path string, data []byte) error {
	d.DryRunInfo.FilesToWrite = append(d.DryRunInfo.FilesToWrite, path)
	return nil
}

func (d *DryRunRecorder) EnsureDirectory(path string) error {
	return nil
}

func (d *DryRunRecorder) Record(key, value string) {
	d.DryRunInfo.Variables[key] = value
}

func NewDryRunRecorder() *DryRunRecorder {
	return &DryRunRecorder{
		DryRunInfo: &DryRunInfo{
			Variables:    make(map[string]string),
			FilesToWrite: make([]string, 0),
		},
	}
}
