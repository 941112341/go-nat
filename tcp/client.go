package tcp

import (
	"errors"
	"log"
	"net"
	"proxy/config"
	"proxy/utils"
	"time"
)


type Client struct {
	Callback func(request *HttpTransferRequest) HttpTransferResponse
	Connection
	AckChannel chan Ack // 通知注册情况
}

func (client *Client) Close() {
	log.Println("close")
	if client.Closed {
		return
	}
	client.CloseLock.Lock()
	defer client.CloseLock.Unlock()
	if client.Closed {
		return
	}
	client.Closed = true
	err := client.Connection.Close()
	if err != nil {
		log.Println(err)
	}
	close(client.ReadChannel)
}



func Connect(ip, port, url string, callback func(request *HttpTransferRequest) HttpTransferResponse) (*Client, error) {
	conn, err := net.Dial("tcp", ip + ":" + port)
	if err != nil {
		return nil, err
	}
	readChannel, ackChannel := make(chan interface{}, 100), make(chan Ack)
	host, err := utils.GetHost()
	if err != nil {
		return nil, err
	}
	client := &Client{
		Connection: Connection{Conn: conn, ReadChannel: readChannel, RegisterRequest: RegisterRequest{
			RegisterURL: url,
			Host:host,
		}},
		AckChannel: ackChannel,
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
	case <-ackChannel:
	case <-time.NewTicker(3 * time.Second).C:
		client.Close()
		return nil, errors.New("register time out")
	}
	log.Println("build success")
	return client, nil
}

func SendHeartBeat(client *Client) {
	defer client.Close()
	for range time.NewTicker(time.Duration(config.HeartBeatRate) * time.Second).C {
		log.Println("send heart beat")
		err := client.WriteHeartBeat()
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func ClientReading(client *Client) {
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

func ReadingListener(client *Client) {
	for key := range client.ReadChannel {
		switch key.(type) {
		case *Ack:
			ack := key.(*Ack)
			if !ack.Ok {
				log.Println(ack.Msg)
				client.Close()
			} else {
				client.AckChannel <- *ack
			}
			log.Println("get register ack")
		case *HttpTransferRequest:
			request := key.(*HttpTransferRequest)
			resp := client.Callback(request)
			err := client.WriteTransferResponse(&resp)
			err = utils.NetErrHandler(err)
			if err != nil {
				log.Println(err)
			}
		default:
			log.Println("err type", key)
		}
	}
}