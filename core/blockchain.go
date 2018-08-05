package core

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"os"
)

// 区块链结构
type BlockChain struct {
	//Blocks []*Block
	tip []byte // 区块链的最后一个区块的Hash
	db  *bolt.DB
}

// MineBlock mines a new block with the provided transactions
// 发送币意味着创建新的交易，并通过挖出新块的方式将交易打包到区块链中
func (bc *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte
	var err error

	bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	newBlock := NewBlock(transactions, lastHash)

	bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(newBlock.Hash, newBlock.Serialize())
		err = b.Put([]byte("l"), newBlock.Hash) // 存储链中最后一个块的哈希
		HandleErr(err)

		bc.tip = newBlock.Hash

		return nil
	})
}

// 迭代器的初始状态为链中的 tip，因此区块将从尾到头（创世块为头）
func (bc *BlockChain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

func (bc *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	// 该地址没有被花掉的交易，意思说交易有输出没有被输入引用
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0 // 总币数

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

func (bc *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		OutPuts:
			for outIdx, out := range tx.Vout {
				// Was the output spent ?
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue OutPuts
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

func (bc *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput

	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

/**
1.打开一个数据库文件
2. 检查文件里面是否已经存储了一个区块链
3. 如果已经存储了一个区块链：
	创建一个新的 Blockchain 实例
	设置 Blockchain 实例的 tip 为数据库中存储的最后一个块的哈希
4. 如果没有区块链：
	创建创世块
	存储到数据库
	将创世块哈希保存为最后一个块的哈希
	创建一个新的 Blockchain 实例，初始时 tip 指向创世块（tip 有尾部，尖端的意思，在这里 tip 存储的是最后一个块的哈希）
*/
func NewBlockChain() *BlockChain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	HandleErr(err)

	// 打开一个 BoltDB 文件的标准做法:这个数据库是key-value形式的。
	// 数据库操作通过一个事务（transaction）进行操作。有两种类型的事务：只读（read-only）和读写（read-write）
	// 打开的是一个读写事务（db.Update(...)），因为我们可能会向数据库中添加创世块
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket)) // 获取区块链数据
		tip = b.Get([]byte("l"))

		return nil
	})

	return &BlockChain{tip, db}
}

func CreateBlockchain(address string) *BlockChain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	HandleErr(err)

	db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTransaction(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		HandleErr(err)

		err = b.Put(genesis.Hash, genesis.Serialize())
		HandleErr(err)

		err = b.Put([]byte("l"), genesis.Hash)
		HandleErr(err)
		tip = genesis.Hash

		return nil
	})

	bc := BlockChain{tip, db}

	return &bc
}

// 关闭db数据库连接
func (bc *BlockChain) DbClose() {
	bc.db.Close()
}
