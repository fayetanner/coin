package core

import (
	"math/big"
	"strconv"
	"bytes"
	"math"
	"fmt"
	"crypto/sha256"
)

// 目前我们并不会实现一个动态调整目标的算法，所以将难度定义为一个全局的常量即可
const targetBits = 24 // 挖矿难度值
const maxNonce = math.MaxInt64 // nonce计算器最大值

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
		pow.block.Data,
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

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		// 找到小于目标targets的哈希值
		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\r%x", hash)
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

// 整数转化为字节数组
func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 10))
}
