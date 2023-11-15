package main

import (
	"bitcoinrpcschema/internal/gensite"
	"log"
	"os"
)

const dbPath = "rpc.db"
const webPath = "www"

func main() {
	db, err := os.ReadFile(dbPath)
	if err != nil {
		log.Fatalln(err)
	}
	err = gensite.Gen(db, webPath)
	if err != nil {
		log.Fatalln(err)
	}
}
