package webkit

import "context"

/*
场景编码
Web: 1 (二进制 001)
Android: 2 (二进制 010)
iOS: 4 (二进制 100)

存储示例
只支持 Web: 1 (二进制 001)
只支持 Android: 2 (二进制 010)
只支持 iOS: 4 (二进制 100)
支持 Web 和 Android: 3 (二进制 011)
支持 Web 和 iOS: 5 (二进制 101)
支持 Android 和 iOS: 6 (二进制 110)
支持所有平台: 7 (二进制 111)

要查询支持特定平台的记录，可以使用位运算。例如，要查询支持 iOS 的记录：
SELECT * FROM your_table WHERE platforms & 4 > 0;
*/

const (
	PlatformWeb     = "web"
	PlatformAndroid = "android"
	PlatformIOS     = "ios"
)

func GetRequestPlatform(ctx context.Context) (string, int) {
	platform := GetHeader(ctx, "platform")
	switch platform {
	case PlatformAndroid:
		return PlatformAndroid, 2
	case PlatformIOS:
		return PlatformIOS, 4
	default:
		// empty string
		return PlatformWeb, 1
	}
}
