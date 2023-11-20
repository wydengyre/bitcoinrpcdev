package bitcoind

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"regexp"
	"strconv"
	"strings"
)

type Command struct {
	Name string
	Help string
}

var versionRe = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

func getVersion(c *rpcclient.Client) (rv ReleaseVersion, err error) {
	i, err := c.GetNetworkInfo()
	if err != nil {
		return
	}
	matches := versionRe.FindStringSubmatch(i.SubVersion)
	if len(matches) < 4 {
		err = fmt.Errorf("could not parse version %s: got regex matches %v", i.SubVersion, matches)
		return
	}

	rv.Major, err = atou(matches[1])
	if err != nil {
		err = fmt.Errorf("could not parse major version from %s: %w", matches[1], err)
		return
	}
	rv.Minor, err = atou(matches[2])
	if err != nil {
		err = fmt.Errorf("could not parse minor version from %s: %w", matches[2], err)
		return
	}
	rv.Patch, err = atou(matches[3])
	if err != nil {
		err = fmt.Errorf("could not parse patch version from %s: %w", matches[3], err)
		return
	}
	return
}

func atou(s string) (uint, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return uint(i), nil
}

func getCommandHelps(c *rpcclient.Client) (map[string][]Command, error) {
	cmds, err := getCommands(c)
	if err != nil {
		return nil, err
	}

	helps := make(map[string][]Command)
	for section, commands := range cmds {
		for _, command := range commands {
			help, err := getCommandHelp(c, command)
			if err != nil {
				return nil, err
			}
			helps[section] = append(helps[section], Command{Name: command, Help: help})
		}
	}
	return helps, nil
}

var commandSectionRe = regexp.MustCompile(`^== (.+) ==$`)

func getCommands(c *rpcclient.Client) (map[string][]string, error) {
	var help string
	resp, err := c.RawRequest("help", nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &help)
	if err != nil {
		return nil, err
	}

	var sectionName string
	var section []string
	commands := make(map[string][]string)
	scanner := bufio.NewScanner(strings.NewReader(help))
	for scanner.Scan() {
		line := scanner.Text()
		matches := commandSectionRe.FindStringSubmatch(line)
		if len(matches) > 1 {
			if len(section) > 0 {
				commands[sectionName] = section
			}
			sectionName = strings.ToLower(matches[1])
			section = make([]string, 0, 1)
		} else if len(line) > 1 && line[0] >= 'a' && line[0] <= 'z' {
			split := strings.Split(line, " ")
			if len(split) == 0 {
				e := fmt.Errorf("could not parse command %s", line)
				return nil, e
			}
			commandName := split[0]
			section = append(section, commandName)
		}
	}
	return commands, nil
}

func getCommandHelp(c *rpcclient.Client, command string) (string, error) {
	var help string
	jsonCommand := json.RawMessage(fmt.Sprintf(`"%s"`, command))
	resp, err := c.RawRequest("help", []json.RawMessage{jsonCommand})
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(resp, &help)
	if err != nil {
		return "", err
	}
	return help, nil
}
