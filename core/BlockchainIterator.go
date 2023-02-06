package core

import "github.com/boltdb/bolt"

// 区块链迭代器
// block chain iterator
type BlockchainIterator struct {
	currentHash []byte   // 当前迭代的块哈希  the hash value of the current interator
	db          *bolt.DB // 区块链指的是存储了一个数据库连接的 Blockchain 实例
	// Blockchain refers to a Blockchain instance that stores a database connection
}

// 只会做一件事情：返回链中的下一个块。
// Will only do one thing: return the next block in the chain.
func (bci *BlockchainIterator) Next() *Block {
	var block *Block

	bci.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get([]byte(bci.currentHash))
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	bci.currentHash = block.PrevBlockHash

	return block
}
