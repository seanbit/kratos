package webkit

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func GenerateUniqueString(size int) string {
	b := make([]byte, size) // Adjust size for your needs.
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	// Get the current timestamp
	timestamp := time.Now().Unix()

	// Combine the timestamp and the random string
	return fmt.Sprintf("%d%s", timestamp, hex.EncodeToString(b))
}

func FirstN(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func AddThousandsSeparatorFromInt(num int64) string {
	numStr := strconv.FormatInt(num, 10)
	if len(numStr) <= 3 {
		return numStr
	}

	// 从末尾开始每三位添加逗号
	result := ""
	for i := len(numStr); i > 0; i -= 3 {
		if i-3 <= 0 {
			result = numStr[:i] + result
		} else {
			result = "," + numStr[i-3:i] + result
		}
	}

	return result
}

func AddThousandsSeparatorFromFloatString(num string) string {
	if num == "" {
		return num
	}

	nums := strings.Split(num, ".")
	n, _ := strconv.ParseInt(nums[0], 10, 64)
	res := AddThousandsSeparatorFromInt(n)
	if len(nums) > 1 {
		res += "." + nums[1]
	}

	return res
}

func KeepDecimalPlaces(num float64, decimal int) string {
	e := int64(1)
	for i := 0; i < decimal; i++ {
		e = e * 10
	}

	n := int64(num * float64(e))
	s := strconv.FormatInt(n, 10)
	if len(s) <= decimal {
		m := decimal - len(s) + 1
		for i := 0; i < m; i++ {
			s = "0" + s
		}
	}

	return s[:len(s)-decimal] + "." + s[len(s)-decimal:]
}
