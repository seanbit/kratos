package service

import (
	"github.com/seanbit/kratos/template/api/web"
	"github.com/seanbit/kratos/template/internal/biz"
)

var (
	blockchainTypes = map[web.BlockChainType]string{
		web.BlockChainType_BLOCK_CHAIN_TYPE_EVM: biz.BlockChainTypeEvm,
	}
)
