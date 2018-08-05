package main

import (
	"coin/core"
	"fmt"
	"time"
)

var startTime = time.Now()

func main() {
	//bc := core.NewBlockChain() // 初始化区块链
	//bc.AddBlock("Send 1 BTC to Amy")
	//bc.AddBlock("Send 10 BTC to Bink")
	//
	//for _,block := range bc.Blocks {
	//	pow := core.NewProofOfWork(block)
	//	fmt.Printf("Prev: %x\n", block.PrevBlockHash)
	//	fmt.Printf("Data: %x\n", block.Data)
	//	fmt.Printf("Hash: %x\n", block.Hash)
	//	fmt.Printf("Pow: %s\n", strconv.FormatBool(pow.Validate()))
	//	fmt.Println()
	//}

	cli := core.CLI{}
	cli.Run()

	fmt.Println("Done use time: ", time.Now().Sub(startTime))
}