package logger

import (
    log "github.com/sirupsen/logrus"
    "github.com/fatih/color"
    "fmt"
)

type PlainFormatter struct {
}

func (f *PlainFormatter) Format(entry *log.Entry) ([]byte, error) {
    cyan := color.New(color.Cyan).SprintFunc()
    return []byte(fmt.Sprintf("%s %s\n",cyan("[Draft]"), entry.Message)), nil
}

