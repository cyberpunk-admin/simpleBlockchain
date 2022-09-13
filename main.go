package main

import (
	"github.com/simpleBlockchain/blockchain"
)

func main() {
	cli := blockchain.NewCLI()
	cli.Run()
}
