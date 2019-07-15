package tcp

import (
	"encoding/json"
	"log"
	"proxy/utils"
	"reflect"
	"strings"
)

// 默认绑定函数
func BinderReq(request *HttpTransferRequest, obj interface{}) {
	err := json.Unmarshal(utils.String2Splice(request.Body), obj)
	if err != nil {
		log.Println(err)
		return
	}
	val := reflect.ValueOf(obj).Elem() // *p
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		name, ok := field.Tag.Lookup("name")
		if !ok {
			name = field.Name
		}
		params := append(request.FormQuery[name], request.Headers[name]...)
		if len(params) == 0 {
			continue
		}
		var param string
		if field.Type.Kind() == reflect.Slice {
			param = "[" + strings.Join(params, ",") + "]"
		} else {
			param = strings.Join(params, ",")
		}
		err := Set(val.Field(i), param)
		if err != nil {
			log.Println(err)
			return
		}
	}
}


func BindResp(obj interface{}) (header map[string]string, body map[string]string) {
	typ := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)
	body = make(map[string]string)
	header = make(map[string]string)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag
		str, ok := tag.Lookup("type")
		if !ok {
			str = "body"
		}
		key, ok := tag.Lookup("key")
		if !ok {
			key = field.Name
		}
		i := reflect.ValueOf(val)
		bytes, err := json.Marshal(i)
		if err != nil {
			log.Println(err)
			continue
		}
		if str == "body" {
			body[key] = string(bytes)
		} else {
			header[key] = string(bytes)
		}
	}
	return
}



// 传进来非指针value
func Set(value reflect.Value, str string) (err error) {
	defer func() {
		if rec, ok := recover().(error); ok {
			err = rec
		}
	}()
	if value.Kind() == reflect.String {
		value.Set(reflect.ValueOf(str))
		return
	}
	i := reflect.New(value.Type()).Interface() // 指针 *p
	err = json.Unmarshal(utils.String2Splice(str), i)
	value.Set(reflect.ValueOf(i).Elem()) // p -> p
	return
}
