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
	log.Println("handler http request")
	requestURI := r.URL.Path
	client := tcp.GetRegisterClient(requestURI)
	if client == nil {
		log.Println("not found client")
		w.WriteHeader(500)
		w.Write(utils.String2Bytes("error"))
		return
	}
	request := tcp.Transmit(r)
	if request == nil {
		log.Println("接受http消息失败")
		w.WriteHeader(500)
		w.Write(utils.String2Bytes("error"))
		return
	}

	err := client.WriteTransferRequest(request)
	if err != nil {
		log.Println("send transfer http request fail" + err.Error())
		w.WriteHeader(500)
		return
	}
	var resp tcp.HttpTransferResponse
	select {
	case resp = <-client.HttpTransferChannel:
	case <-time.NewTicker(time.Second).C:
		log.Println("请求超时")
		w.WriteHeader(500)
		return
	}
	for key, value := range resp.Headers {
		w.Header().Add(key, value)
	}
	w.WriteHeader(200)
	_, err = w.Write(utils.String2Bytes(resp.Body))
	if err != nil {
		log.Println("write http response fail", err)
		client.Close()
		tcp.RemoveRegister(client)
	}
}

func Start() {
	http.HandleFunc("/", Handler)
	_ = http.ListenAndServe(":"+config.HttpPort, nil)
}
