package spinner

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/briandowns/spinner"
)

// type Spinner struct {
// 	spinner *spinner.Spinner
// }

var s *spinner.Spinner

func init() {
	s = createSpinner("--> Setting up Github OIDC...")
}

func createSpinner(msg string) *spinner.Spinner {
	cyan := color.New(color.Bold, color.FgCyan).SprintFunc()
	s = spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Prefix = fmt.Sprintf("%s %s ", cyan("[Draft]"), msg)
	s.Suffix = " "
	return s
}

func GetSpinner() *spinner.Spinner {
	if s == nil {
		s = createSpinner("--> Setting up Github OIDC...")
	}
	
	return s
}