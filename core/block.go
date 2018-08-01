package core

import (
	"time"
)

// 区块结构体定义
type Block struct {
	Timestamp     int64 // 区块创建时间戳
	Data          []byte // 区块包含的数据
	PrevBlockHash []byte // 前一个区块的哈希值
	Hash          []byte // 区块自身的哈希值，用于校验区块数据有效
	Nonce         int // 工作量证明难度值,用来校验数据的
}

// NewBlock create and return Block
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:time.Now().UnixNano(),
		Data: []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash: []byte{},
	}

	// 采用工作量证明得出的新区块
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash
	block.Nonce = nonce

	return block
}

// 设置区块自身的Hash值
//func (b *Block) SetHash()  {
//	timestamp := IntToHex(b.Timestamp)
//	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
//	hash := sha256.Sum256(headers)
//	b.Hash = hash[:]
//}

// 创建创世区块
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}