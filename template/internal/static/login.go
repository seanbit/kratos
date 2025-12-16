package static

const (
	LoginDomain                    = "index.unicornx.ai"
	LoginSignaturePrefixTextFormat = "Hello! %s asks you to sign this message to confirm your ownership of the address. This action will not cost any gas fee."
	LoginSignatureTextFormat       = "%s \n\nHere is your account: %s\n\nExpiration time: %s\n"
	// LoginSignatureTextPattern 正则表达式匹配模板格式
	// 第一个%s是任意字符（非贪婪），第二个%s是Here is your account后的内容，第三个是时间字符串
	LoginSignatureTextPattern = `(?s).*?\n\nHere is your account: (.*?)\n\nExpiration time: (.*?)\n`
)
