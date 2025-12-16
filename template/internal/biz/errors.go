package biz

import (
	"github.com/seanbit/kratos/template/api/web"
)

var (
	ErrSignatureTextInvalid       = web.ErrorAuthSignatureTextInvalid("Invalid signature text format!")
	ErrSignatureTextExpired       = web.ErrorAuthSignatureTextExpired("Signature text Expired!")
	ErrBlockChainTypeNotSupported = web.ErrorAuthBlockChainTypeNotSupport("blockchain type not supported")
	ErrLoginExpired               = web.ErrorAuthLoginExpired("login expired")
	ErrLoginTokenInvalid          = web.ErrorAuthLoginTokenInvalid("login token invalid")

	ErrUserNotFound = web.ErrorUserNotFound("user not found")
)
