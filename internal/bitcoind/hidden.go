package bitcoind

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"slices"
)

func getHiddenCommands(dir string) ([]string, error) {
	regHeaderPath := path.Join(dir, "register.h")
	reg, err := os.ReadFile(regHeaderPath)
	if err != nil {
		e := fmt.Errorf("failed to read file: %w", err)
		return nil, e
	}
	registrations, err := getRegistrations(reg)
	if err != nil {
		e := fmt.Errorf("failed to get registrations: %w", err)
		return nil, e
	}

	ps, err := os.ReadDir(dir)
	if err != nil {
		e := fmt.Errorf("failed to read dir: %w", err)
		return nil, e
	}

	var cmds []string
	for _, p := range ps {
		if p.IsDir() {
			continue
		}
		if path.Ext(p.Name()) != ".cpp" {
			continue
		}

		f, err := os.Open(path.Join(dir, p.Name()))
		if err != nil {
			e := fmt.Errorf("failed to open file: %w", err)
			return nil, e
		}
		// errors are fine, as we don't expect all CPP files to have hidden commands
		hidden, _ := getHiddenCmds(f, registrations)
		closeErr := f.Close()
		if closeErr != nil {
			e := fmt.Errorf("failed to close file: %w", closeErr)
			return nil, e
		}
		cmds = append(cmds, hidden...)
	}

	return cmds, nil
}

var rpcSectionRe = regexp.MustCompile(`(?m)void (Register[A-Z]\w+RPCCommands)\(.+\);$`)

func getRegistrations(regFileContent []byte) ([]string, error) {
	matches := rpcSectionRe.FindAllSubmatch(regFileContent, -1)
	registrations := make([]string, len(matches))
	for i, match := range matches {
		if len(match) < 2 {
			return nil, fmt.Errorf("invalid match %v: %v", i, match)
		}
		registrations[i] = string(match[1])
	}
	return registrations, nil
}

var registerFnRe = regexp.MustCompile(`void (Register\w+RPCCommands)\(.*\)$`)
var hiddenCmdRe = regexp.MustCompile(`^\s+{"hidden", &(.+)},$`)

func getHiddenCmds(cmdFileContent io.Reader, fnNames []string) ([]string, error) {
	var cmds []string
	scanner := bufio.NewScanner(cmdFileContent)
	found := false
	for scanner.Scan() {
		line := scanner.Bytes()
		match := registerFnRe.FindSubmatch(line)
		if len(match) < 2 {
			continue
		}
		fnName := string(match[1])
		if slices.Contains(fnNames, fnName) {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("failed to find cmd registrations for cmds %v", fnNames)
	}

	for scanner.Scan() {
		line := scanner.Bytes()
		match := hiddenCmdRe.FindSubmatch(line)
		if len(match) < 2 {
			continue
		}
		cmdName := string(match[1])
		cmds = append(cmds, cmdName)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan cmd registrations: %w", err)
	}
	return cmds, nil
}
