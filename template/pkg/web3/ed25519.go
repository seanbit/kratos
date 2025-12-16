package web3

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"github.com/pkg/errors"
)

type Ed25519KeyPairEncode = string

const (
	Ed25519KeyPairEncodeHex Ed25519KeyPairEncode = "hex"
	Ed25519KeyPairEncodeB64 Ed25519KeyPairEncode = "base64"
)

type Ed25519KeyPair struct {
	Key    string               `json:"key"`
	Pub    string               `json:"pub"`
	Encode Ed25519KeyPairEncode `json:"encode"`
}

func GenEd25519KeyPair(encode Ed25519KeyPairEncode) (*Ed25519KeyPair, error) {
	// 生成 Ed25519 密钥对
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate keys")
	}

	keyPair := &Ed25519KeyPair{Encode: encode}
	switch encode {
	case Ed25519KeyPairEncodeHex:
		keyPair.Pub = hex.EncodeToString(publicKey)
		keyPair.Key = hex.EncodeToString(privateKey)
	case Ed25519KeyPairEncodeB64:
		keyPair.Pub = base64.StdEncoding.EncodeToString(publicKey)
		keyPair.Key = base64.StdEncoding.EncodeToString(privateKey)
	default:
		return nil, errors.Errorf("Unsupported Ed25519 key pair encode type: %s", encode)
	}
	return keyPair, nil
}

// LoadEd25519Keys 解析私钥并生成 Ed25519 私钥和公钥
func LoadEd25519Keys(privateKeyHex string, encode Ed25519KeyPairEncode) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	var (
		privateKeyBytes []byte
		err             error
	)
	// 解码十六进制私钥
	switch encode {
	case Ed25519KeyPairEncodeHex:
		privateKeyBytes, err = hex.DecodeString(privateKeyHex)
	case Ed25519KeyPairEncodeB64:
		privateKeyBytes, err = base64.StdEncoding.DecodeString(privateKeyHex)
	default:
		return nil, nil, errors.New("Unsupported Ed25519 key type")
	}
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to load private key")
	}

	// 检查私钥长度是否正确
	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return nil, nil, errors.New("invalid private key length")
	}

	// 将私钥转换为 ed25519.PrivateKey 类型
	privateKey := ed25519.PrivateKey(privateKeyBytes)

	// 从私钥中提取公钥
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return privateKey, publicKey, nil
}
