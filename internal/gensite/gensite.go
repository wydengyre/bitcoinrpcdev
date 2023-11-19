package gensite

import (
	"bitcoinrpcschema/internal/bitcoind"
	_ "embed"
	"fmt"
	"slices"
)

//go:embed pico.min.css
var picoCss []byte

func Gen(db []byte, webPath string) error {
	rpcDb, err := bitcoind.ReadDb(db)
	if err != nil {
		return err
	}

	site := newSite()
	for rv, sections := range rpcDb {
		for sec, cmds := range sections {
			for _, cmd := range cmds {
				p := fmt.Sprintf("%s/%s/%s/index.html", rv, sec, cmd.Name)
				c := &command{
					Version:     rv.String(),
					Section:     sec,
					Name:        cmd.Name,
					Description: cmd.Help,
				}
				err := site.add(p, c)
				if err != nil {
					return fmt.Errorf("failed to add command %s to site: %w", cmd.Name, err)
				}
			}
			p := fmt.Sprintf("%s/%s/index.html", rv, sec)
			s := &section{
				Name:     sec,
				Version:  rv.String(),
				Commands: cmdNames(cmds),
			}
			err := site.add(p, s)
			if err != nil {
				return fmt.Errorf("failed to add section %s to site: %w", sec, err)
			}
		}
		name := rv.String()
		p := name + "/index.html"
		sections := cmdNamesBySection(sections)
		v := version{name, sections}
		err := site.add(p, &v)
		if err != nil {
			return fmt.Errorf("failed to add version %s to site: %w", rv.String(), err)
		}
	}

	idx := &index{}
	idx.Latest, idx.Versions, err = versionsDescending(rpcDb)
	if err != nil {
		return fmt.Errorf("failed to get versions: %w", err)
	}
	err = site.add("index.html", idx)
	if err != nil {
		return fmt.Errorf("failed to add index to site: %w", err)
	}

	site.addRaw("pico.min.css", picoCss)

	err = site.write(webPath)
	if err != nil {
		return fmt.Errorf("failed to write site: %w", err)
	}
	return nil
}

func cmdNames(cmds []bitcoind.Command) []string {
	cmdNames := make([]string, len(cmds))
	for i, cmd := range cmds {
		cmdNames[i] = cmd.Name
	}
	return cmdNames
}

func cmdNamesBySection(commands map[string][]bitcoind.Command) map[string][]string {
	sections := make(map[string][]string)
	for section, cmds := range commands {
		for _, cmd := range cmds {
			sections[section] = append(sections[section], cmd.Name)
		}
	}
	return sections
}

func versionsDescending(db bitcoind.RpcDb) (string, []string, error) {
	if len(db) < 1 {
		return "", nil, fmt.Errorf("empty database")
	}
	versions := make([]bitcoind.ReleaseVersion, 0, len(db))
	for v := range db {
		versions = append(versions, v)
	}
	slices.SortFunc(versions, func(a, b bitcoind.ReleaseVersion) int {
		return -a.Cmp(b)
	})
	latest := versions[0].String()
	other := make([]string, len(versions)-1)
	for i, v := range versions[1:] {
		other[i] = v.String()
	}
	return latest, other, nil
}
