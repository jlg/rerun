package exec

import (
	"fmt"
	"os"
	osexec "os/exec"
)

type Partial struct {
	osexec.Cmd
}

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

func (p *Partial) Run(args ...string) error {
	cmd := p.Cmd // Make a copy of command before applying remaining args
	cmd.Args = append(cmd.Args, args...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("exec %s: %w", cmd.String(), err)
	}
	return nil
}
