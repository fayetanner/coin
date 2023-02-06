package core

// 交易输出结构体
// transaction output struct
type TXOutput struct {
	Value        int    // 一定量的比特币(Value)
	ScriptPubKey string // 一个锁定脚本(ScriptPubKey)，要花这笔钱，必须要解锁该脚本。
	// To spend this value, the ScriptPubKey must be unlocked.
}

// 判断该输入是否可以被unlockingData解锁
// Determine whether the input can be unlocked by unlockingData
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}
