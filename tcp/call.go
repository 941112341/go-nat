package tcp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"proxy/utils"
	"sync"
	"time"
	"unsafe"
)

// open -> close -> remove
type ServerClient struct {
	Connection
	HttpTransferChannel chan HttpTransferResponse
	LastHeartBeat time.Time
}


type Connection struct {
	net.Conn
	RegisterRequest
	Closed bool // 避免重复关闭
	CloseLock sync.Mutex
	ReadChannel chan interface{}
}

func (client *ServerClient) Close() {
	p(errors.New("client close"), client)
	if client.Closed {
		return
	}
	client.CloseLock.Lock()
	defer client.CloseLock.Unlock()
	if client.Closed {
		return
	}
	client.Closed = true
	if err := client.Connection.Close(); err != nil {
		log.Println(err)
	}
	close(client.HttpTransferChannel)
	close(client.ReadChannel)
}

func (client *Connection) WriteTransferRequest(request *HttpTransferRequest) error {
	return client.writeMsg(TRANSFER_REQ, request)
}

func (client *Connection) WriteTransferResponse(response *HttpTransferResponse) error {
	return client.writeMsg(TRANSFER_RESP, response)
}

func (client *Connection) WriteACK(ack *Ack) error {
	return client.writeMsg(ACK, ack)
}

func (client *Connection) WriteRegister() error {
	return client.writeMsg(REGISTER, client.RegisterRequest)
}

func (client *Connection) WriteHeartBeat() error {
	return client.writeMsg(HEART_BEAT, HeartBeat{})
}

func (client *Connection) writeMsg(mtype MessageType, protocal interface{}) error {
	bytes, err := json.Marshal(protocal)
	if err != nil {
		return nil
	}
	msg := &Message{
		MType:   mtype,
		Content: *(*string)(unsafe.Pointer(&bytes)),
	}
	return client.WriteMessage(msg)
}

func (client *Connection) WriteMessage(message *Message) error {
	msg := message.ToString()
	bytes := utils.String2Splice(msg)
	n, err := client.Write(bytes)
	if err != nil { // 超时
		log.Println(fmt.Errorf("tcp:write timeout %v", err))
		return err
	}
	for n < len(bytes) {
		bytes = bytes[n:]
		n, err = client.Write(bytes)
		if err != nil { // 超时
			log.Println(fmt.Errorf("tcp:write timeout %v", err))
			return err
		}
	}
	return nil
}

func (client *Connection) ReadACK() (*Ack, error) {
	msg, err := client.ReadMessage()
	if err != nil {
		return nil, err
	}
	var ack Ack
	err = json.Unmarshal(utils.String2Splice(msg.Content), &ack)
	return &ack, err
}

func (client *Connection) ReadRegister() (*RegisterRequest, error) {
	msg, err := client.ReadMessage()
	if err != nil {
		return nil, err
	}
	var ack RegisterRequest
	err = json.Unmarshal(utils.String2Splice(msg.Content), &ack)
	return &ack, err
}

func (client *Connection) ReadTransferRequest() (*HttpTransferRequest, error) {
	msg, err := client.ReadMessage()
	if err != nil {
		return nil, err
	}
	var ack HttpTransferRequest
	err = json.Unmarshal(utils.String2Splice(msg.Content), &ack)
	return &ack, err
}

func (client *Connection) ReadTransferResponse() (*HttpTransferResponse, error) {
	msg, err := client.ReadMessage()
	if err != nil {
		return nil, err
	}
	var ack HttpTransferResponse
	err = json.Unmarshal(utils.String2Splice(msg.Content), &ack)
	return &ack, err
}


func (client *Connection) ReadHeartBeat() (*HeartBeat, error) {
	msg, err := client.ReadMessage()
	if err != nil {
		return nil, err
	}
	var ack HeartBeat
	err = json.Unmarshal(utils.String2Splice(msg.Content), &ack)
	return &ack, err
}


func (client *Connection) ReadMessage() (msg *Message, err error) {
	// 超时控制
	//err = client.SetReadDeadline(time.Now().Add(timeout))
	//if err != nil {
	//	return
	//}
	r := bufio.NewReader(client)
	msgType, err := r.ReadString('\n')
	if err != nil {
		return
	}
	msgContent, err := r.ReadString('\n')
	if err != nil {
		return
	}
	msg = NewMessage(msgType + msgContent)
	if msg == nil {
		err = fmt.Errorf("read message 失败")
	}
	return
}

