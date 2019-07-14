package tcp

import (
	"errors"
	"fmt"
	"log"
	"net"
	"proxy/config"
	"proxy/http"
	"time"
)

const (
	maxConnect = 500
)

func Start() {
	listen, err := net.Listen("tcp", ":"+config.TcpPort)
	if err != nil {
		log.Println(err)
	}
	limiter := NewLimiter(maxConnect)
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
		}
		limiter.Fuse()
		go handler(conn)
	}
}

func handler(conn net.Conn) {
	channel := make(chan HttpTransferResponse)
	msgChannel := make(chan interface{}, 100)
	client := &ServerClient{
		Connection:          Connection{Conn: conn, ReadChannel:msgChannel},
		HttpTransferChannel: channel,
	}
	request, err := client.ReadRegister()
	if err != nil {
		p(err, client)
		client.Close()
		return
	}
	client.RegisterRequest = *request
	ok := http.AddRegister(client)
	if !ok {
		err = client.WriteFail("路径已经注册")
	} else {
		err = client.WriteSuccess()
	}
	if err != nil {
		p(err, client)
		client.Close()
		return
	}
	// 持续读
	go ServerReading(client)
	go ListenMessage(client)
}

func ListenMessage(client *ServerClient) {
	for msg := range client.ReadChannel {
		switch msg.(type) {
		case HeartBeat:
			client.LastHeartBeat = time.Now()
		case HttpTransferResponse:
			select {
			case client.HttpTransferChannel <- msg.(HttpTransferResponse):
			case <-time.NewTicker(time.Second).C:
				p(fmt.Errorf("入队失败 %#v", msg), client)
			}
		default:
			p(errors.New("error msg type"), client)
		}
	}
}


func (client *ServerClient) WriteSuccess() error {
	return client.WriteACK(&Ack{
		Ok: true,
	})
}

func (client *ServerClient) WriteFail(msg string) error {
	return client.WriteACK(&Ack{
		Msg: msg,
	})
}


func ServerReading(client *ServerClient) {
	defer client.Close()
	for {
		msg, err := client.ReadMessage()
		if err != nil {
			return
		}
		client.ReadChannel <- msg
	}
}

func p(err error, client *ServerClient) {
	log.Println(client.RegisterURL, client.Host, err.Error())
}