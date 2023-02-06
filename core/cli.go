package core

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// cli命令常量列表
// cli command constant list
const (
	cliGetBalance       = "getbalance"
	cliCreateBlockchain = "createblockchain"
	cliSend             = "send"
	cliPrintChain       = "printchain"
)

// cli命令结构体
// cli command struct
type CLI struct {
	//bc *BlockChain
}

// 启动cli命令
// cli run command
func (cli *CLI) Run() {
	var err error
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet(cliGetBalance, flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet(cliCreateBlockchain, flag.ExitOnError)
	sendCmd := flag.NewFlagSet(cliSend, flag.ExitOnError)
	printChainCmd := flag.NewFlagSet(cliPrintChain, flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	// 解析命令行参数
	// Parse command line arguments
	switch os.Args[1] {
	case cliGetBalance:
		err = getBalanceCmd.Parse(os.Args[2:])
		HandleErr(err)
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)

	case cliCreateBlockchain:
		err = createBlockchainCmd.Parse(os.Args[2:])
		HandleErr(err)
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)

	case cliPrintChain:
		err = printChainCmd.Parse(os.Args[2:])
		HandleErr(err)
		cli.printChain()

	case cliSend:
		err = sendCmd.Parse(os.Args[2:])
		HandleErr(err)
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)

	default:
		cli.printUsage()
		os.Exit(1)
	}
}

// 检查命令行参数：至少要有两个
// check command line arguments: there should be minimun two.
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// 打印命令使用说明
// Display command line usage instructions
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

// 添加一个新区块
// add a new block
//func (cli *CLI) addBlock(data string) {
//	cli.bc.AddBlock(data)
//	fmt.Println("Add New Block Success!")
//}

// 创建区块链，创世区块
// create block chain, genesis block
func (cli *CLI) createBlockchain(address string) {
	bc := CreateBlockchain(address)
	defer bc.db.Close()
	fmt.Println("Done!")
}

// 打印区块链，从最新块开始->创世区块
// Print the blockchain, starting from the latest block -> genesis block
func (cli *CLI) printChain() {
	bc := NewBlockChain()
	defer bc.db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

// 获取余额
// obtain balance
func (cli *CLI) getBalance(address string) {
	bc := NewBlockChain()
	defer bc.DbClose()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d BTC\n", address, balance)
}

// 转账(即是转币)
// send coin
func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockChain()
	defer bc.DbClose()

	// 创建转账交易记录
	// Create transfer transaction records
	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})

	fmt.Println("Success!")
}
