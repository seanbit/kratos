package loghelper

import (
	"fmt"
	"strings"
)

func FieldErr(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf(", error: %+v", err)
}

func String(k, v string) string {
	return fmt.Sprint(", ", k, ": ", v)
}

func Bool(k string, v bool) string {
	return fmt.Sprint(", ", k, ": ", v)
}

func Int(k string, v int) string {
	return fmt.Sprint(", ", k, ": ", v)
}

func Int32(k string, v int32) string {
	return fmt.Sprint(", ", k, ": ", v)
}

func Int64(k string, v int64) string {
	return fmt.Sprint(", ", k, ": ", v)
}

func FieldMod(value string) string {
	value = strings.Replace(value, " ", ".", -1)
	return fmt.Sprint(", mod: ", value)
}

func FieldKey(value string) string {
	return fmt.Sprint(", key: ", value)
}

func Any(key string, value interface{}) string {
	return fmt.Sprint(", ", key, ": ", value)
}

func ByteString(key string, value []byte) string {
	return fmt.Sprint(", ", key, ": ", string(value))
}
