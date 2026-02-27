package check

import (
	"context"
	"fmt"
	"os/exec"
)

type BinaryCheck struct {
	Binary string
}

func (c *BinaryCheck) Name() string {
	return fmt.Sprintf("%s installed", c.Binary)
}

func (c *BinaryCheck) Run(_ context.Context) Result {
	_, err := exec.LookPath(c.Binary)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("%s not found on PATH", c.Binary),
			Fix:     fmt.Sprintf("Install %s and make sure it is on your PATH", c.Binary),
		}
	}
	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("%s is installed", c.Binary),
	}
}
