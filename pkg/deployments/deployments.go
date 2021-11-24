package deployments

import (
	"embed"
	"fmt"
	"github.com/imiller31/draftv2/pkg/embedutils"
	"github.com/imiller31/draftv2/pkg/osutil"
	"io/fs"
	"log"
)

//go:generate cp -r ../../deployTypes ./deployTypes

var (
	//go:embed deployTypes
	deployTypes embed.FS
	parentDirName = "deployTypes"
)

type Deployments struct {
	deploys map[string]fs.DirEntry
	dest string
}

func (d *Deployments) CopyDeploymentFiles(deployType string) error {
	val, ok := d.deploys[deployType]
	if !ok {
		return fmt.Errorf("deployment type: %s is not currently supported", deployType)
	}

	srcDir := parentDirName + "/" + val.Name()

	if err := osutil.CopyDir(deployTypes, srcDir, d.dest); err != nil {
		return err
	}

	return nil
}



func CreateDeployments() *Deployments {
	deployMap, err := embedutils.EmbedFStoMap(deployTypes, "deployTypes")
	if err != nil {
		log.Fatal(err)
	}

	return &Deployments{
		deploys: deployMap,
		dest: "./",
	}
}