package spinner

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/briandowns/spinner"
)

var s *spinner.Spinner

func init() {
	CreateSpinner("--> Setting up Github OIDC...")
}

func CreateSpinner(msg string) *spinner.Spinner {
	cyan := color.New(color.Bold, color.FgCyan).SprintFunc()
	s = spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Prefix = fmt.Sprintf("%s %s ", cyan("[Draft]"), msg)
	s.Suffix = " "
	return s
}

func GetSpinner() *spinner.Spinner {
	return s
}