// Package exec provides osexec.Cmd with partial args.
package exec

import (
	"fmt"
	"os"
	osexec "os/exec"
)

// Partial implements osexec.Cmd with part of arguments applied.
type Partial struct {
	osexec.Cmd
}

// NewParial creates new Partial with os.Stdout and os.Stderr and initial args.
func NewPartial(args ...string) *Partial {
	p := Partial{}
	if len(args) == 0 {
		p.Err = fmt.Errorf("no executable file provided")
		return &p
	}
	p.Cmd = *osexec.Command(args[0], args[1:]...)
	p.Stdout = os.Stdout
	p.Stderr = os.Stderr
	return &p
}

// Run finalizes the partial command but appending reminder args to a Cmd
// and Runs the finalized Cmd.
func (p *Partial) Run(args ...string) error {
	cmd := p.Finalize(args...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("exec %s: %w", cmd.String(), err)
	}
	return nil
}

// Finalize finalizes the Parital by copying embedded Cmd and appending
// remaining args to the Cmd.
func (p *Partial) Finalize(args ...string) *osexec.Cmd {
	cmd := p.Cmd // Make a copy of command before applying remaining args
	cmd.Args = append(cmd.Args, args...)
	return &cmd
}
