package utils

import (
	"time"
	"unsafe"
)

func String2Bytes(str string) []byte {
	return *(*[]byte)(unsafe.Pointer(&str))
}

func Bytes2String(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

func RandomInt64() int64 {
	return time.Now().UnixNano()
}

func Now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}