package bitcoind

import (
	"bytes"
	_ "embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

//go:embed test/register.h
var registerH []byte

//go:embed test/mining.cpp
var miningCpp []byte

func TestGetRegistrations(t *testing.T) {
	registrations, err := getRegistrations(registerH)
	require.NoError(t, err)
	expected := []string{"RegisterBlockchainRPCCommands", "RegisterFeeRPCCommands", "RegisterMempoolRPCCommands", "RegisterMiningRPCCommands", "RegisterNodeRPCCommands", "RegisterNetRPCCommands", "RegisterOutputScriptRPCCommands", "RegisterRawTransactionRPCCommands", "RegisterSignMessageRPCCommands", "RegisterSignerRPCCommands", "RegisterTxoutProofRPCCommands"}
	assert.Equal(t, expected, registrations)
}

func TestGetHiddenCmds(t *testing.T) {
	r := bytes.NewReader(miningCpp)
	registrations, err := getHiddenCmds(r, []string{"RegisterMiningRPCCommands"})
	require.NoError(t, err)
	expected := []string{"generatetoaddress", "generatetodescriptor", "generateblock", "generate"}
	assert.Equal(t, expected, registrations)
}
