package main

import (
	"bitcoinrpcschema/internal/downloader"
	"log"
)

const rootPath = "bitcoin-core"

func main() {
	err := downloader.Get(rootPath)
	if err != nil {
		log.Fatalln(err)
	}
}
