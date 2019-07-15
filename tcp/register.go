package tcp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"proxy/config"
	"sync"
	"time"
	"unsafe"
)

// 注册表
var RegisterMap = map[string]*ServerClient{} // URL -> client
var lock sync.Mutex

func AddRegister(client *ServerClient) bool  {
	lock.Lock()
	defer lock.Unlock()
	if rc, ok := RegisterMap[client.RegisterURL]; !ok {
		client.LastHeartBeat = time.Now()
		RegisterMap[client.RegisterURL] = client
		p(errors.New("添加client注册"), client)
		return true
	} else {
		if rc.Host == client.Host {
			RegisterMap[client.RegisterURL] = client
			return true
		} else {
			return false
		}
	}
}

func RemoveRegister(client *ServerClient) {
	lock.Lock()
	defer lock.Unlock()
	delete(RegisterMap, client.RegisterURL)
	p(errors.New("删除client注册"), client)
}

func GetRegisterClient(url string) *ServerClient {
	lock.Lock()
	defer lock.Unlock()
	rc, ok := RegisterMap[url]
	if !ok {
		return nil
	}
	return rc
}

// http -> 转发
func Transmit(r *http.Request) *HttpTransferRequest {
	request := &HttpTransferRequest{}
	request.Method = r.Method
	request.Headers = r.Header
	err := r.ParseForm()
	if err != nil {
		request.ProxyMsg = err.Error()
	}
	request.FormQuery = r.Form

	length := r.ContentLength
	bytes := make([]byte, length)
	i, err := r.Body.Read(bytes)

	if err != nil {
		// EOF -> 请求方主动关闭了socket
		if err != io.EOF {
			log.Println(err)
			return nil
		} else {
			log.Println("EOF")
		}
	} else if int64(i) < length { // socket不满，不会等待，非阻塞，这里不会有这个问题
		log.Println(fmt.Sprintf("http:transmit: read less than contentLength %#v", request))
		panic("没有读满整个http体")
	}
	request.Body = *(*string)(unsafe.Pointer(&bytes))
	return request
}

func init() {
	go CheckDeadConnect()
}

func CheckDeadConnect() {
	for range time.NewTicker(time.Duration(config.RegisterTimeout) * time.Second).C {
		func() {
			lock.Lock()
			defer lock.Unlock()
			for _, value := range RegisterMap {
				if subTime := time.Now().Sub(value.LastHeartBeat); subTime > time.Duration(config.RegisterTimeout) * time.Second {
					delete(RegisterMap, value.RegisterURL)
					p(errors.New("删除过期client"), value)
					value.Close()
				}
			}
		} ()
	}
}