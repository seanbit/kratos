package cryptos

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

func GetStringFromEnv(k string) string {
	v := os.Getenv(k)
	if len(v) == 0 {
		panic(fmt.Sprintf("%s NOT FOUND", k))
	}
	return v
}

// pkcs7Pad 对数据进行 PKCS7 填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpad 去除 PKCS7 填充
func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("invalid padding: empty data")
	}
	padding := int(data[length-1])
	if padding > length || padding > aes.BlockSize || padding == 0 {
		return nil, fmt.Errorf("invalid padding size")
	}
	for i := 0; i < padding; i++ {
		if data[length-1-i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}
	return data[:length-padding], nil
}

// EncryptByKey aesKey: AES密钥，16个字符
func EncryptByKey(aesKey string, text string) (string, error) {
	if len(aesKey) != 16 {
		return "", fmt.Errorf("wrong aesKey")
	}
	key := []byte(aesKey)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %v", err)
	}

	// 使用 PKCS7 填充明文
	paddedPlaintext := pkcs7Pad(plaintext, aes.BlockSize)

	// 使用CBC模式加密
	ciphertext := make([]byte, aes.BlockSize+len(paddedPlaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate random iv: %v", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], paddedPlaintext)

	// 返回Base64编码的密文
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func DecryptByKey(aesKey string, ciphertext string) (string, error) {
	key := []byte(aesKey)
	encrypted, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %v", err)
	}

	// 检查密文长度是否合法
	if len(encrypted) < aes.BlockSize*2 {
		return "", fmt.Errorf("invalid ciphertext length")
	}

	// 检查密文长度是否为块大小的倍数
	if (len(encrypted)-aes.BlockSize)%aes.BlockSize != 0 {
		return "", fmt.Errorf("invalid ciphertext: not a multiple of block size")
	}

	// 提取iv
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	// 使用CBC模式解密
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encrypted, encrypted)

	// 去除 PKCS7 填充数据并返回明文
	plaintext, err := pkcs7Unpad(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to unpad: %v", err)
	}

	return string(plaintext), nil
}
