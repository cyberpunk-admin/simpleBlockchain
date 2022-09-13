package blockchain

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	createblockchain = "createblockchain"
	send             = "send"
	printchain       = "printchain"
	printblance      = "printbalance"
	createwallet     = "createwallet"
)

type CLI struct{}

func NewCLI() CLI {
	return CLI{}
}

func (cli CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet(createblockchain, flag.ExitOnError)
	sendCmd := flag.NewFlagSet(send, flag.ExitOnError)
	printChainCmd := flag.NewFlagSet(printchain, flag.ExitOnError)
	printBalanceCmd := flag.NewFlagSet(printblance, flag.ExitOnError)
	createwalletCmd := flag.NewFlagSet(createwallet, flag.ExitOnError)

	sendFrom := sendCmd.String("from", "", "send coin address from")
	sendTo := sendCmd.String("to", "", "send coin address to")
	sendAmount := sendCmd.Int("amount", 0, "send coin amount")

	createBlockchainArr := createBlockchainCmd.String("address", "", "the address of reward sent to for create genesis block")
	printBlanceArr := printBalanceCmd.String("address", "", "the address of wallet")

	switch os.Args[1] {
	case createblockchain:
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case send:
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case printchain:
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case printblance:
		err := printBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case createwallet:
		err := createwalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainArr == "" {
			cli.printUsage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainArr)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			cli.printUsage()
			os.Exit(1)
		}
		cli.sent(*sendFrom, *sendTo, *sendAmount)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if printBalanceCmd.Parsed() {
		if *printBlanceArr == "" {
			cli.printUsage()
			os.Exit(1)
		}
		cli.printBalance(*printBlanceArr)
	}

	if createwalletCmd.Parsed() {
		cli.createWallet()
	}
}

func (cli *CLI) createBlockchain(address string) {
	CreateBlockchain(address)
	fmt.Println("blockchain created!")
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  send -Send some coins from a address to another address")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  printbalance - Print a address's balance")
	fmt.Println("  createwallet - create a wallet and print it address")
}

func (cli *CLI) printChain() {
	bc := GetBlockchain()
	defer bc.Close()

	it := bc.Iterator()
	for {
		block := it.Next()

		fmt.Printf("PervHash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("Pow: %v\n\n", pow.Validate())

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) printBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: address is not valid")
	}

	bc := GetBlockchain()
	defer bc.Close()

	balance := 0
	pubKeyHash := getPubKeyHashFromAddress(address)

	UTXOs := bc.FindUTXOs(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) sent(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := GetBlockchain()

	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	txs := NewUTXOTransaction(&wallet, to, amount, bc)

	bc.MineBlock([]*Transaction{txs})
	fmt.Printf("success send %s to %s %v coins\n", from, to, amount)
}

func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()

	wallets.SaveToFile()

	fmt.Printf("a new addres: %s\n", address)
}
