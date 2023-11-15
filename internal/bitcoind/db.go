package bitcoind

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path"
)

type RpcDb map[ReleaseVersion]map[string][]Command

// this is also defined in downloader, but the two needn't be identical
type ReleaseVersion struct {
	Major uint
	Minor uint
	Patch uint
}

func (v ReleaseVersion) String() string {
	if v.Patch != 0 {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

func CreateDb(daemonPath string) ([]byte, error) {
	rpcDb, err := mkDb(daemonPath)
	if err != nil {
		e := fmt.Errorf("error getting commands: %v", err)
		return nil, e
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err = enc.Encode(rpcDb)
	if err != nil {
		e := fmt.Errorf("error encoding commands: %v", err)
		return nil, e
	}
	return b.Bytes(), nil
}

func ReadDb(db []byte) (RpcDb, error) {
	var cmds RpcDb
	dec := gob.NewDecoder(bytes.NewReader(db))
	err := dec.Decode(&cmds)
	if err != nil {
		e := fmt.Errorf("error decoding commands: %v", err)
		return nil, e
	}
	return cmds, nil
}

func mkDb(daemonPath string) (RpcDb, error) {
	dirs, err := os.ReadDir(daemonPath)
	if err != nil {
		e := fmt.Errorf("error reading directory: %v", err)
		return nil, e
	}

	db := make(RpcDb, len(dirs))
	for _, dir := range dirs {
		entryPath := path.Join(daemonPath, dir.Name())
		version, cmds, err := getRpcInfo(entryPath)
		if err != nil {
			log.Printf("error getting commands: %v", err)
		}
		db[version] = cmds
	}
	return db, nil
}

func getRpcInfo(versionPath string) (ReleaseVersion, map[string][]Command, error) {
	bitcoindPath := path.Join(versionPath, "bin", "bitcoind")
	return GetDaemonCommands(bitcoindPath)
}

func GetDaemonCommands(bitcoindPath string) (ReleaseVersion, map[string][]Command, error) {
	conf, err := startBitcoind(bitcoindPath)
	if err != nil {
		return ReleaseVersion{}, nil, err
	}
	defer conf.Cleanup()
	c := conf.Client

	v, err := getVersion(c)
	if err != nil {
		e := fmt.Errorf("error getting version: %w", err)
		return ReleaseVersion{}, nil, e
	}

	cmds, err := getCommandHelps(c)
	if err != nil {
		e := fmt.Errorf("error getting commands: %w", err)
		return ReleaseVersion{}, nil, e
	}
	return v, cmds, nil
}
