package core

import (
	"math/big"
	"bytes"
	"fmt"
	"crypto/sha256"
)

// 工作量证明结构
type ProofOfWork struct {
	block *Block
	// 我们会将哈希与目标进行比较：先把哈希转换成一个大整数，然后检测它是否小于目标。
	target *big.Int // 目标(target)的指针
}

func NewProofOfWork(b *Block) *ProofOfWork {
	// 我们将 big.Int 初始化为 1，然后左移 256 - targetBits 位。
	// 256 是一个 SHA-256 哈希的位数，我们将要使用的是 SHA-256 哈希算法
	// target（目标） 的 16 进制形式为：
	// 0x10000000000000000000000000000000000000000000000000000000000
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

// 准备数据进行哈希运算 nonce: Hashcash计数器
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.HashTransaction(),
		IntToHex(pow.block.Timestamp),
		IntToHex(int64(targetBits)),
		IntToHex(int64(nonce)),
	}, []byte{})

	return data
}

// 运行工作量证明得出新的区块哈希值以及Nonce
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining a new block \n")
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		// 找到小于目标targets的哈希值
		fmt.Printf("\r%x", hash)
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// 工作量生成的区块哈希值验证是否是有效的
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
