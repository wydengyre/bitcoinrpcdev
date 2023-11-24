package gensite

import (
	"bufio"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed command.html
var commandHtml string

type command struct {
	Version     string
	Section     string
	Name        string
	DateTime    string
	Description string
}

type parsedDescription struct {
	Usage       string
	Explanation []string
	Arguments   string
	Result      string
}

type commandTmplData struct {
	Command           *command
	ParsedDescription *parsedDescription
}

var commandTmpl = mustBtcTemplate("command", commandHtml)

func (c *command) html() ([]byte, error) {
	desc, err := parseDescription(c.Description)
	if err != nil {
		e := fmt.Errorf("failed to parse description: %w", err)
		return nil, e
	}

	ctd := commandTmplData{
		Command:           c,
		ParsedDescription: desc,
	}

	rendered, err := commandTmpl.render(ctd)
	if err != nil {
		e := fmt.Errorf("failed to render command %s html: %w", c.Name, err)
		return nil, e
	}
	return rendered, nil
}

func parseDescription(description string) (*parsedDescription, error) {
	p := &parsedDescription{}
	s := bufio.NewScanner(strings.NewReader(description))

	s.Scan()
	p.Usage = s.Text()

	skipArguments := false
	for s.Scan() {
		t := s.Text()
		if t == "Arguments:" {
			break
		}
		if strings.HasPrefix(t, "Result") {
			skipArguments = true
			break
		}
		if len(t) > 0 {
			p.Explanation = append(p.Explanation, s.Text())
		}
	}

	b := strings.Builder{}
	if !skipArguments {
		for s.Scan() {
			t := s.Text()
			if strings.HasPrefix(t, "Result") {
				break
			}
			b.WriteString(t)
			b.WriteString("\n")
		}
		p.Arguments = b.String()
	}

	b = strings.Builder{}
	for s.Scan() {
		b.WriteString(s.Text())
		b.WriteString("\n")
	}
	p.Result = b.String()

	return p, s.Err()
}
