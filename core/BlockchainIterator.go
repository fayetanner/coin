package core

import "github.com/boltdb/bolt"

// 区块链迭代器
type BlockchainIterator struct {
	currentHash []byte   // 当前迭代的块哈希
	db          *bolt.DB // 区块链指的是存储了一个数据库连接的 Blockchain 实例
}

// 只会做一件事情：返回链中的下一个块。
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
