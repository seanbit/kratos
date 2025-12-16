package web3

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

type EthereumAccount struct {
	PrivateKey    *ecdsa.PrivateKey
	PrivateKeyHex string
	Address       common.Address
	AddressHex    string
}

func GenerateEthereumAccount() (*EthereumAccount, error) {
	// 1. 创建私钥
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, errors.Wrap(err, "生成私钥失败")
	}

	// 2. 从私钥获取地址
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// 3. 显示私钥（十六进制格式）
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)[2:] // 去掉0x前缀

	return &EthereumAccount{
		PrivateKey:    privateKey,
		PrivateKeyHex: privateKeyHex,
		Address:       address,
		AddressHex:    address.Hex(),
	}, nil
}
