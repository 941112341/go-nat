package tcp

import (
	"errors"
	"reflect"
)

// fun 是一个函数
func Initialize(ip, port, url string, i interface{}) (client *Client, err error) {
	typ := reflect.TypeOf(i)
	if typ.Kind() != reflect.Func {
		return nil, errors.New("参数不是函数")
	}
	reflectHandler := func(request *HttpTransferRequest) HttpTransferResponse {


	}
	return Connect(ip, port, url, reflectHandler)
}
