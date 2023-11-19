package main

import (
	"bitcoinrpcschema/internal/purgecss"
	"log"
	"os"
)

func main() {
	cssPath := os.Args[1]
	htmlPath := os.Args[2]
	err := purgecss.PurgeSite(cssPath, htmlPath)
	if err != nil {
		log.Fatalln(err)
	}
}
