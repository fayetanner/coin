package core

import (
	"log"
	"os"
	"strconv"
)

// 整数转化为字节数组
// Integer converted to byte array
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
// Check if the file or folder exists
// Use the error value returned by the os.Stat() function to determine:
// If the returned error is nil, the file or folder exists
// If the returned error type is determined to be true by os.IsNotExist(), it means that the file or folder does not exist
// If the returned error is other type, it is not sure whether it exists
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
