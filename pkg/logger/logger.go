package logger

import (
	"bytes"
	"fmt"
	"os"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

type CustomFormatter struct {}

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
    cyan := color.New(color.Bold, color.FgCyan).SprintFunc()
    return []byte(fmt.Sprintf("%s %s\n",cyan("[Draft]"), entry.Message)), nil
}

type OutputSplitter struct{}

func (splitter *OutputSplitter) Write(p []byte) (n int, err error) {
	if bytes.Contains(p, []byte("level=error")) ||  bytes.Contains(p, []byte("level=fatal")) || bytes.Contains(p, []byte("level=panic")) {
		return os.Stderr.Write(p)
	}
	return os.Stdout.Write(p)
}