package logger

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	cyan := color.New(color.Bold, color.FgCyan).SprintFunc()
	red := color.New(color.Bold, color.FgRed).SprintFunc()
	level := strings.Title(entry.Level.String())
	if level == "Error" || level == "Fatal" || level == "Panic" {
		return []byte(fmt.Sprintf("%s: %s\n", red(level), entry.Message)), nil
	}
	return []byte(fmt.Sprintf("%s %s\n", cyan("[Draft]"), entry.Message)), nil
}

type OutputSplitter struct{}

func (splitter *OutputSplitter) Write(p []byte) (n int, err error) {
	if bytes.Contains(p, []byte("Error")) || bytes.Contains(p, []byte("Fatal")) || bytes.Contains(p, []byte("Panic")) {
		return os.Stderr.Write(p)
	}
	return os.Stdout.Write(p)
}
