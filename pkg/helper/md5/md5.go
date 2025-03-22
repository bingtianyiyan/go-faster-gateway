package md5

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func MD5(data []byte) string {
	// 创建 MD5 哈希对象
	hash := md5.New()
	// 将字符串转换为字节数组并写入哈希对象
	hash.Write(data)
	// 计算哈希值
	hashedBytes := hash.Sum(nil)
	// 将哈希值转换为十六进制字符串表示
	md5Hash := hex.EncodeToString(hashedBytes)
	return md5Hash
}
