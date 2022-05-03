package logger

import (
    log "github.com/sirupsen/logrus"
    "github.com/fatih/color"
    "fmt"
)

type CustomFormatter struct {}

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
    cyan := color.New(color.Bold, color.FgCyan).SprintFunc()
    return []byte(fmt.Sprintf("%s %s\n",cyan("[Draft]"), entry.Message)), nil
}
