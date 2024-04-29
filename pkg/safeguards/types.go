package safeguards

import "io/fs"

type FileCrawler struct {
	Safeguards   []Safeguard
	constraintFS fs.FS
}

type Safeguard struct {
	name           string
	templatePath   string
	constraintPath string
}

type ManifestFile struct {
	Name string
	Path string
}

type ManifestResult struct {
	Name             string              // the name of the manifest
	ObjectViolations map[string][]string // a map of string object names to slice of string objectViolations
}
