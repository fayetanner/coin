package core

// 区块链结构
type BlockChain struct {
	Blocks []*Block
}

// 添加一个新的区块到区块链
func (bc *BlockChain) AddBlock(data string) {
	prevBlock := bc.Blocks[len(bc.Blocks) -1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.Blocks = append(bc.Blocks, newBlock)
}

// 创建区块链，并创建创世区块放进区块链
func NewBlockChain() *BlockChain {
	return &BlockChain{[]*Block{NewGenesisBlock()}}
}
