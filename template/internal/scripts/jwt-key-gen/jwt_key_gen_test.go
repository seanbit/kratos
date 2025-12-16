package main

import (
	"fmt"
	"testing"

	"github.com/seanbit/kratos/template/pkg/web3"
)

func TestParseJwtKey(t *testing.T) {
	printKeyHex := "7ff75f302acc68b7e0bac1423d0e94956f1118e568e93df65383e6dceea1316284f5165f688ad4e1c6bcbc1355ea62e9c29085e13be4eba9b2291120ca59899e"
	// 从十六进制私钥生成 Ed25519 私钥和公钥
	privateKey, publicKey, err := web3.LoadEd25519Keys(printKeyHex, web3.Ed25519KeyPairEncodeHex)
	if err != nil {
		fmt.Printf("Failed to load keys: %v\n", err)
		return
	}

	// 输出私钥和公钥
	fmt.Printf("Private Key: %x\n", privateKey)
	fmt.Printf("Public Key: %x\n", publicKey)
}
