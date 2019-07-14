package http

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"proxy/tcp"
	"sync"
	"time"
	"unsafe"
)

// 注册表
var RegisterMap map[string]*tcp.ServerClient // URL -> client
var lock sync.Mutex

func AddRegister(client *tcp.ServerClient) bool  {
	lock.Lock()
	defer lock.Unlock()
	if rc, ok := RegisterMap[client.RegisterURL]; !ok {
		RegisterMap[client.RegisterURL] = client
		return true
	} else {
		return rc.Host == client.Host
	}
}

func RemoveRegister(client *tcp.ServerClient) {
	lock.Lock()
	defer lock.Unlock()
	delete(RegisterMap, client.RegisterURL)
}

func GetRegisterClient(url string) *tcp.ServerClient {
	lock.Lock()
	defer lock.Unlock()
	rc, ok := RegisterMap[url]
	if !ok {
		return nil
	}
	return rc
}

// http -> 转发
func Transmit(r *http.Request) *tcp.HttpTransferRequest {
	request := &tcp.HttpTransferRequest{}
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
			request.ProxyMsg = err.Error()
		}
	} else if int64(i) < length { // socket不满，不会等待，非阻塞，这里不会有这个问题
		log.Println(fmt.Sprintf("http:transmit: read less than contentLength %#v", request))
	}
	request.Body = *(*string)(unsafe.Pointer(&bytes))
	return request
}

func init() {
	go CheckDeadConnect()
}

func CheckDeadConnect() {
	for range time.NewTicker(10 * time.Second).C {
		func() {
			lock.Lock()
			defer lock.Unlock()
			for _, value := range RegisterMap {
				if subTime := time.Now().Sub(value.LastHeartBeat); subTime > 5*time.Second {
					delete(RegisterMap, value.RegisterURL)
					value.Close()
				}
			}
		} ()
	}
}