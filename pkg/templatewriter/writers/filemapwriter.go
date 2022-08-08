package writers

type FileMapWriter struct {
	FileMap map[string][]byte
}

func (w *FileMapWriter) WriteFile(path string, data []byte) error {
	if w.FileMap == nil {
		w.FileMap = map[string][]byte{}
	}

	w.FileMap[path] = data
	return nil
}

func (w *FileMapWriter) EnsureDirectory(path string) error {
	return nil
}
