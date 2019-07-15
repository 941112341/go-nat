package utils

import (
	"errors"
	"io"
	"log"
	"net"
	"reflect"
	"unsafe"
)

func String2Bytes(str string) []byte { // 之后这样是无法切片的
	return *(*[]byte)(unsafe.Pointer(&str))
}

func String2Splice(str string) []byte {/**/
	var sh reflect.SliceHeader
	h := *(*reflect.StringHeader)(unsafe.Pointer(&str))
	sh.Data, sh.Cap, sh.Len = h.Data, h.Len, h.Len
	return *(*[]byte)(unsafe.Pointer(&sh))
}

func GetHost() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(),nil
			}
		}
	}
	return "", errors.New("not found ip v4")
}

func NetErrHandler(err error) error {
	if err == io.EOF {
		return err // 客户端关闭了直接返回
	}
	switch e := err.(type) {
	case net.Error:
		if e.Timeout() || e.Temporary() {
			log.Println(e)
			return nil
		}
		if opErr, ok := e.(*net.OpError); ok {
			log.Println(opErr)
		}
	case nil:
		return nil
	default:
		log.Println("未知异常", err)
	}
	return err
}