package bootstrap

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/DATA-DOG/godog"
)

type commandContext struct {
}

func RegisterCommandContext(s *godog.Suite) {
	cc := &commandContext{}

	s.Step(`^I run a command "([^"]*)" with args "([^"]*)":$`, cc.iRunACommand)
}

func (c *commandContext) iRunACommand(command, args string) error {
	cArgs := strings.Split(args, " ")
	var out bytes.Buffer

	cmd := exec.Command(command, cArgs...)
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("running command %s. Error: %s\n%s\n", command, err.Error(), out.String())
	}

	return nil
}
