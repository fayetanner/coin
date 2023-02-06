package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// 交易结构体，用来存储一笔交易
// transaction struct, used to store a transaction
type Transaction struct {
	ID []byte // 交易id,使用输入输出等信息来哈希，确保信息不被篡改
	// transaction id, Use information such as
	//input and output to hash to ensure that the information is not tampered with
	Vin  []TXInput
	Vout []TXOutput
}

// 设置交易的ID编号，这里是做hash处理
// Set the ID number of the transaction, which is processed with hash algorithm
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	HandleErr(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// IsCoinbase check whether the transaction is coinbase
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// 创建挖矿奖励交易：交易只有一个输出，没有输入
// Create a mining reward transaction: the transaction has only one output and no input
func NewCoinbaseTransaction(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := TXInput{[]byte{}, -1, data} // -1表示该输入没有引用任何输出
	// -1 means that the input does not refer to any output
	txout := TXOutput{subsidy, to} // 一个块给的奖励subsidy = 10
	// the reward subsidy given by/to a block is 10
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

// 创建转账交易记录
// Create transfer transaction records
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	// 找到可以用来花费的所有有效交易输出(即是统计出该原地址的所有的币)
	// Find all valid transaction outputs that can be used for
	// spending (that is, count all the coins of the original address)
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	// 判断该地址from的币是否够用来该笔转账
	// Determine whether the coins from the address "from" is enough for the transfer
	if acc < amount {
		log.Panic("Error: Not enough funds to spend.")
	}

	// build a list inputs for this transaction
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		HandleErr(err)
		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	// build a list of outputs for this transaction
	// 转账输出给(to)
	// transaction output to "to"
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		// 找零输出,输出给原账户(from)
		// change output, given to original account "from"
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
