package main

import (
	"bitcoinrpcschema/internal/downloader"
	"log"
)

const rootPath = "bitcoin-core"

const binUrl = "https://bitcoincore.org/bin/"

const gitUrl = "https://github.com/bitcoin/bitcoin.git"

func main() {
	err := downloader.Get(rootPath, binUrl, gitUrl)
	if err != nil {
		log.Fatalln(err)
	}
}
