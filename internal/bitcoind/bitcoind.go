package bitcoind

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type Config struct {
	Client  *rpcclient.Client
	Cleanup func()
}

func startBitcoind(path string) (conf Config, err error) {
	var tmpDirectory string
	tmpDirectory, err = os.MkdirTemp("", "bitcoinrpcschema-bitcoind")
	if err != nil {
		return
	}

	removeTempDir := func() {
		removeErr := os.RemoveAll(tmpDirectory)
		if removeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to remove tmp directory: %v\n", removeErr)
		}
	}

	cmd := exec.Command(path, "-server", "-regtest", "-daemonwait", "-datadir="+tmpDirectory)
	if err = cmd.Start(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			removeTempDir()
		}
	}()
	err = cmd.Wait()
	if err != nil {
		return
	}

	pidPath := tmpDirectory + "/regtest/bitcoind.pid"
	pidBytes, err := os.ReadFile(pidPath)
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return
	}

	cleanup := func() {
		err := p.Signal(syscall.SIGTERM)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to stop bitcoind: %v\n", err)
		} else {
			_, _ = p.Wait()
		}
		removeTempDir()
	}

	networkParams := chaincfg.RegressionNetParams
	host := "localhost:18443" // regtest default
	cookiePath := tmpDirectory + "/regtest/.cookie"
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         host,
		Params:       networkParams.Name,
		DisableTLS:   true,
		HTTPPostMode: true,
		CookiePath:   cookiePath,
	}, nil)
	if err != nil {
		cleanup()
		return
	}
	conf = Config{Client: client, Cleanup: cleanup}
	return
}
