package cryptos

import (
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

	// 使用CBC模式加密，填充明文数据
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate random iv: %v", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

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
	if len(encrypted) < aes.BlockSize {
		return "", fmt.Errorf("invalid ciphertext length")
	}

	// 提取iv
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	// 使用CBC模式解密
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encrypted, encrypted)

	// 去除填充数据并返回明文
	plaintext := make([]byte, len(encrypted))
	copy(plaintext, encrypted)

	return string(plaintext), nil
}
