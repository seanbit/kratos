package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/seanbit/kratos/template/internal/biz"
	"github.com/seanbit/kratos/template/pkg/web3"
)

func TestAuth_WalletSignVerify(t *testing.T) {

	auth := &biz.Auth{}
	evmAccount, err := web3.GenerateEthereumAccount()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.TODO()

	t.Run("AuthVerifySignatureNotExpired", func(t *testing.T) {
		testSignatureTextExpiresDuration := time.Second * 200
		message := auth.GetLoginSignatureText(ctx, biz.BlockChainTypeEvm, evmAccount.AddressHex, testSignatureTextExpiresDuration)
		fmt.Println(message)
		signature, err := web3.SignatureEthereumMessage(ctx, message, evmAccount.PrivateKey)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(signature)
		if err := auth.VerifyLoginSignature(ctx, biz.BlockChainTypeEvm, message, signature, evmAccount.AddressHex); err != nil {
			t.Error(err)
		}
	})
	t.Run("AuthVerifySignatureHasExpired", func(t *testing.T) {
		testSignatureTextExpiresDuration := time.Second * 2
		message := auth.GetLoginSignatureText(ctx, biz.BlockChainTypeEvm, evmAccount.AddressHex, testSignatureTextExpiresDuration)
		fmt.Println(message)
		time.Sleep(time.Second * 3)
		signature, err := web3.SignatureEthereumMessage(ctx, message, evmAccount.PrivateKey)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(signature)
		if err := auth.VerifyLoginSignature(ctx, biz.BlockChainTypeEvm, message, signature, evmAccount.AddressHex); err != nil {
			t.Log(err)
		} else {
			t.Error("expected error but got nil")
		}
	})
}
