package web3

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

func SignatureEthereumMessage(ctx context.Context, text string, privateKey *ecdsa.PrivateKey) (string, error) {
	// 1. 将消息转换为以太坊签名消息格式
	fullMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(text), text)

	// 2. 计算消息的Keccak256哈希
	messageHash := crypto.Keccak256Hash([]byte(fullMessage))

	// 3. 使用私钥对哈希进行签名
	signatureBytes, err := crypto.Sign(messageHash.Bytes(), privateKey)
	if err != nil {
		return "", err
	}
	// 4. 将签名转换为十六进制字符串
	return hexutil.Encode(signatureBytes), nil
}

func VerifyEthereumSignature(ctx context.Context, text string, signatureStr string, checkAddress string) error {
	signature, err := hexutil.Decode(signatureStr)
	if err != nil {
		return err
	}
	if signature == nil || len(signature) < 65 {
		return errors.New("signature is invalid")
	}

	// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md
	if signature[64] >= 35 {
		signature[64] = (signature[64] - 35) % 2
	} else {
		if signature[64] == 27 || signature[64] == 28 {
			signature[64] -= 27
		}
	}

	hash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(text), text)))
	publicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return err
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	if strings.ToLower(checkAddress) != strings.ToLower(address) {
		return errors.New(fmt.Sprintf("signature to address %s is not %s", address, checkAddress))
	}

	return nil
}
