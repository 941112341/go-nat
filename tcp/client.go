package tcp

import (
	"errors"
	"log"
	"net"
	"time"
)


type Client struct {
	Callback func(request *HttpTransferRequest) HttpTransferResponse
	Connection
	Closed bool // 避免重复关闭
	ResponseChannel chan HttpTransferResponse
}

func (client *Client) Close() {
	if client.Closed {
		return
	}
	for err := client.Connection.Close(); err != nil; err = client.Connection.Close() {
		log.Println(err)
	}
	client.Closed = true
	close(client.ReadChannel)
}



func Connect(ip, port, url string, callback func(request *HttpTransferRequest) HttpTransferResponse) (*Client, error) {
	conn, err := net.Dial("tcp", ip + ":" + port)
	readChannel, respChannel := make(chan interface{}, 100), make(chan HttpTransferResponse)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Connection: Connection{Conn: conn, ReadChannel: readChannel, RegisterRequest: RegisterRequest{
			RegisterURL: url,
		}},
		ResponseChannel:respChannel,
		Callback:callback,
	}
	err = client.WriteRegister()
	if err != nil {
		return nil, err
	}
	go SendHeartBeat(client)
	go ClientReading(client)
	go ReadingListener(client)
	select {
	case <-respChannel:
	case <-time.NewTicker(3 * time.Second).C:
		client.Close()
		return nil, errors.New("register time out")
	}
	return client, nil
}

func SendHeartBeat(client *Client) {
	defer client.Close()
	for range time.NewTicker(5 * time.Second).C {
		err := client.WriteHeartBeat()
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func ClientReading(client *Client) {
	defer client.Close()
	for  {
		msg, err := client.ReadMessage()
		if err != nil {
			return
		}
		client.ReadChannel <- msg
	}
}

func ReadingListener(client *Client) {
	for key := range client.ReadChannel {
		switch key.(type) {
		case Ack:
			ack := key.(*Ack)
			if !ack.Ok {
				log.Println(ack.Msg)
				client.Close()
			}
		case HttpTransferRequest:
			request := key.(*HttpTransferRequest)
			resp := client.Callback(request)
			select {
			case client.ReadChannel <- resp:
			case <-time.NewTicker(2 * time.Second).C:
				log.Println("time out")
			}
		default:
			log.Println("err type", key)
		}
	}
}