package blockchain

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	createblockchain = "createblockchain"
	addblock	= "addblock"
	printchain = "printchain"
)

type CLI struct {
	bc *Blockchain
}

func (cli CLI) validateArgs()  {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet(createblockchain, flag.ExitOnError)
	addBlockCmd := flag.NewFlagSet(addblock, flag.ExitOnError)
	printChainCmd := flag.NewFlagSet(printchain, flag.ExitOnError)
	addBlockData := addBlockCmd.String("data", "", "add Block data")

	switch os.Args[1] {
	case createblockchain:
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case addblock:
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case printchain:
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		cli.createBlockchain()
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			cli.printUsage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func (cli *CLI) createBlockchain() {
	bc := NewBlockchain()
	defer bc.db.Close()
	fmt.Println("blockchain created!")
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  addblock -data Add a block with data to blockchain")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
}

func (cli *CLI) printChain()  {
	bc := NewBlockchain()
	defer bc.db.Close()

	it := bc.Iterator()
	for {
		block := it.Next()

		fmt.Printf("PervHash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("Pow: %v\n\n", pow.Validate())

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) addBlock(data string){
	cli.bc =  NewBlockchain()
	defer cli.bc.db.Close()

	err := cli.bc.AddBlock(data)
	if err != nil {
		fmt.Printf("add block failed %s\n", err)
		return
	}
	fmt.Println("add block Success!")
	return
}