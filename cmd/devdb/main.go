package devdb

import (
	"bitcoinrpcschema/internal/bitcoind"
	"log"
	"os"
)

const dbPath = "rpc.dev.db"

const bitcoindPath = "bitcoind"

func main() {
	db, err := bitcoind.CreateDb(daemonPath)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.WriteFile(dbPath, db, 0644)
	if err != nil {
		log.Fatalln(err)
	}

}
