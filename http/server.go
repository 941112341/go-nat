package http

import (
	"log"
	"net/http"
	"proxy/tcp"
	"proxy/utils"
	"time"
)
import "proxy/config"

func Handler(w http.ResponseWriter, r *http.Request) {
	requestURI := r.RequestURI
	client := GetRegisterClient(requestURI)
	if client == nil {
		log.Println("not found client")
		w.WriteHeader(500)
		return
	}
	request := Transmit(r)
	err := client.WriteTransferRequest(request)
	if err != nil {
		log.Println("send transfer http request fail" + err.Error())
		w.WriteHeader(500)
		client.Close()
		RemoveRegister(client)
		return
	}
	var resp *tcp.HttpTransferResponse
	select {
	case resp = <-client.HttpTransferChannel:
	case <-time.NewTicker(time.Second).C:
		log.Println("请求超时")
		w.WriteHeader(500)
		return
	}
	w.Header() = resp.Headers
	w.WriteHeader(200)
	_, err = w.Write(utils.String2Bytes(resp.Body))
	if err != nil {
		log.Println("write http response fail", err)
		client.Close()
		RemoveRegister(client)
	}
}

func Start() {
	http.HandleFunc("/", Handler)
	_ = http.ListenAndServe(config.HttpPort, nil)
}
