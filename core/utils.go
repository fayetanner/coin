package core

import (
	"strconv"
	"log"
)

// 整数转化为字节数组
func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 10))
}

func HandleErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}