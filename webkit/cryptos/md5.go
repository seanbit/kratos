package cryptos

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5EncodeBytesToHex(v []byte) string {
	m := md5.New()
	m.Write(v)
	return hex.EncodeToString(m.Sum(nil))
}

// MD5EncodeStringToHex 计算字符串的MD5哈希值
func MD5EncodeStringToHex(s string) string {
	return MD5EncodeBytesToHex([]byte(s))
}
