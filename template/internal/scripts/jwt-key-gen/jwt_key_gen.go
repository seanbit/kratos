package main

import (
	"fmt"

	"github.com/seanbit/kratos/template/pkg/web3"
)

func main() {
	// 生成 Ed25519 密钥对
	keypair, err := web3.GenEd25519KeyPair(web3.Ed25519KeyPairEncodeHex)
	if err != nil {
		fmt.Printf("Failed to generate keys: %v\n", err)
		return
	}
	// 输出私钥和公钥
	fmt.Printf("Private Key (Hex): %s\n", keypair.Key)
	fmt.Printf("Public Key (Hex): %s\n", keypair.Pub)
}
