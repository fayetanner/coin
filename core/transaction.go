package core

import (
	"log"
	"encoding/hex"
	"fmt"
	"bytes"
	"encoding/gob"
	"crypto/sha256"
)

// 交易结构体，用来存储一笔交易
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// 设置交易的ID编号，这里是做hash处理
func (tx *Transaction) SetID()  {
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
func NewCoinbaseTransaction(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()
	
	return &tx
}

// 创建转账交易记录
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	// 找到可以用来花费的所有有效交易输出
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	// 判断该地址from的币是否够用来该笔转账
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
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		// 找零输出,输出给原账户(from)
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
