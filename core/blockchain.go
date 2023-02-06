package core

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

// 区块链结构
// struct for the blockchain
type BlockChain struct {
	//Blocks []*Block
	tip []byte // 区块链的最后一个区块的Hash  // the hash for the last block in the block chain
	db  *bolt.DB
}

// MineBlock mines a new block with the provided transactions
// 发送币意味着创建新的交易，并通过挖出新块的方式将交易打包到区块链中
// Sending coins means creating a new transaction and
// packing the transaction into the blockchain by mining a new block
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
		// Stores the hash of the last block in the chain
		HandleErr(err)

		bc.tip = newBlock.Hash

		return nil
	})
}

// 迭代器的初始状态为链中的 tip，因此区块将从尾到头（创世块为头）
// The initial state of the iterator is the tip in the chain,
//
//	so the blocks will go from the end to the beginning (the genesis block is the head)
func (bc *BlockChain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// 可以用来spent的输出
// Spendable outputs
func (bc *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	// 该地址没有被花掉的交易，意思说交易有输出没有被输入引用
	// There is no spent transaction at this address,
	// which means that the transaction has an output that is not referenced by an input
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0 // 总币数 total coin

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

// 找到所有包含未花费输出的交易集合
// Find the set of all transactions that contain unspent outputs
func (bc *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID) // 交易ID转成字符串使用
			// encode the transaction ID into string

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

			// 判断是否是创世区块
			// determine if it is genesis block
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		// 表示已经到创世区块了，结束查找
		// genesis block found, end the search.
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// 统计出address对应的所有没有花费出去的输出(也即是它的币)
// Count all unspent outputs corresponding to address (that is, its coins)
func (bc *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput

	// 所有包含未花费输出的交易
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

// 判断db数据库是否存在，也就是判断文件是否存在
// Determine whether the db database exists, that is,
// determine whether the file exists
func dbExists() bool {
	flag, _ := PathExists(dbFile)

	return flag
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

/*
*
1. Open a database file
2. Check if a blockchain is already stored in the file
3. If a blockchain is already stored:
Create a new Blockchain instance
Set the Blockchain instance's tip to the hash of the last block stored in the database
 4. If there is no blockchain:
    Create Genesis Block
    store in database
    Save the genesis block hash as the hash of the last block
    Create a new Blockchain instance, initially the tip points to the genesis block
    (tip has a tail, meaning the tip, here the tip stores the hash of the last block)
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
	/* Standard practice for opening a BoltDB file: the database is in key-value format.
	       Database operations are performed through a transaction.
		   There are two types of transactions: read-only and read-write:
		   What is opened is a read-write transaction (db.Update(...)),
		   because we may add a genesis block to the database
	*/
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket)) // 获取区块链数据 // obtain blockchain data
		tip = b.Get([]byte("l"))             // 最后一个区块的哈希值 // hash value of the last block

		return nil
	})

	return &BlockChain{tip, db}
}

// 创建区块链，即是初始化区块链，添加创世区块
// Creating a blockchain means initializing the blockchain and adding the genesis block
func CreateBlockchain(address string) *BlockChain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	HandleErr(err)

	db.Update(func(tx *bolt.Tx) error {
		// 创世区块交易，只有输出，没有输入
		// Genesis block transaction, only output, no input
		cbtx := NewCoinbaseTransaction(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		HandleErr(err)

		err = b.Put(genesis.Hash, genesis.Serialize())
		HandleErr(err)

		err = b.Put([]byte("l"), genesis.Hash) // 最新块哈希值 // Hash value of the genesis hash
		HandleErr(err)
		tip = genesis.Hash

		return nil
	})

	bc := BlockChain{tip, db}

	return &bc
}

// 关闭db数据库连接
// close the connection to the db database
func (bc *BlockChain) DbClose() {
	bc.db.Close()
}
