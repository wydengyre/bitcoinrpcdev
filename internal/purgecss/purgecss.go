package purgecss

import (
	_ "embed"
	"errors"
	"fmt"
	"os/exec"
)

//go:embed purgecss.mjs
var purgeCSS string

func purgecss(html, css string) ([]byte, error) {
	cmd := nodeCommand(purgeCSS, html, css)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		ok := errors.As(err, &exitErr)
		if ok {
			e := fmt.Errorf("purgecss.mjs stderr %s: %w", exitErr.Stderr, exitErr)
			return nil, e
		}
		e := fmt.Errorf("error running purgecss.mjs: %w", err)
		return nil, e
	}

	return out, nil
}

func nodeCommand(script string, arg ...string) *exec.Cmd {
	nodeArgs := append([]string{"--input-type=module", "--eval=" + script, "--"}, arg...)
	return exec.Command("node", nodeArgs...)
}
