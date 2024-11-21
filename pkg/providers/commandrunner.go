package providers

import (
	"errors"
	"os/exec"
)

// CommandRunner is an interface for executing commands and getting the output/error
type CommandRunner interface {
	RunCommand(...string) (string, error)
}

type DefaultCommandRunner struct{}
var _ CommandRunner = &DefaultCommandRunner{}

func (d *DefaultCommandRunner) RunCommand(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

type FakeCommandRunner struct {
	Output string
	ErrStr string
}
var _ CommandRunner = &FakeCommandRunner{}

func (f *FakeCommandRunner) RunCommand(args ...string) (string, error) {
	if f.ErrStr != "" {
		return f.Output, errors.New(f.ErrStr)
	}
	return f.Output, nil
}
