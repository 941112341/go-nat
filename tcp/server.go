package tcp

import (
	"errors"
	"fmt"
	"log"
	"net"
	"proxy/config"
	"proxy/utils"
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
	log.Println("handler tcp connection")
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
	ok := AddRegister(client)
	if !ok {
		err = client.WriteFail("路径已经注册")
		RemoveRegister(client)
		client.Close()
	} else {
		err = client.WriteSuccess()
	}
	if err != nil {
		log.Println(err)
	}
	// 持续读
	go ServerReading(client)
	go ListenMessage(client)
}

func ListenMessage(client *ServerClient) {
	for msg := range client.ReadChannel {
		switch msg.(type) {
		case *HeartBeat:
			client.LastHeartBeat = time.Now()
			p(errors.New("heart beat"), client)
		case *HttpTransferResponse:
			select {
			case client.HttpTransferChannel <- *msg.(*HttpTransferResponse):
				p(errors.New("http response"), client)
			case <-time.NewTicker(time.Second).C:
				p(fmt.Errorf("转发超时 %#v", msg), client)
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
			err = utils.NetErrHandler(err)
			if err != nil {
				return
			}
			continue
		}
		client.ReadChannel <- msg.ProtoType()
	}
}

func p(err error, client *ServerClient) {
	log.Println(client.RegisterURL, client.Host, err.Error())
}