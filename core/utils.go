package core

import (
	"log"
	"strconv"
	"os"
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

// 判断文件或者文件夹是否存在
// 使用os.Stat()函数返回的错误值进行判断:
// 如果返回的错误为nil,说明文件或文件夹存在
// 如果返回的错误类型使用os.IsNotExist()判断为true,说明文件或文件夹不存在
// 如果返回的错误为其它类型,则不确定是否在存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
