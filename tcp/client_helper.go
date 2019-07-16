package tcp

import (
	"encoding/json"
	"errors"
	"reflect"
)

// fun 是一个函数
func Initialize(ip, port, url string, i interface{}) (client *Client, err error) {
	typ := reflect.TypeOf(i)
	if typ.Kind() != reflect.Func {
		return nil, errors.New("参数不是函数")
	}
	pNumIn := typ.NumIn()
	iarr := make([]reflect.Value, pNumIn)
	reflectHandler := func(request *HttpTransferRequest) HttpTransferResponse {
		for i := 0; i < pNumIn; i++ {
			ptyp := typ.In(i)
			pinst := reflect.New(ptyp).Interface()
			BinderReq(request, pinst)
			iarr[i] = reflect.ValueOf(pinst).Elem()
		}
		ret := reflect.ValueOf(i).Call(iarr)
		// 只取第一个参数
		val := ret[0]
		h, b := BindResp(val.Interface())
		body, _ := json.Marshal(b)
		return HttpTransferResponse{
			Headers: h, Body: string(body),
		}
	}
	return Connect(ip, port, url, reflectHandler)
}
