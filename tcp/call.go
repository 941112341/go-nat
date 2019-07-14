package tcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
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
	ReadChannel chan interface{}
}

func (client *ServerClient) Close() {
	for err := client.Connection.Close(); err != nil; err = client.Connection.Close() {
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
	content := message.Content
	bytes := *(*[]byte)(unsafe.Pointer(&content))
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
	var imsg interface{} = msg
	ack, ok := imsg.(*Ack)
	if !ok {
		return nil, fmt.Errorf("cannot cast message to ack")
	}
	return ack, nil
}

func (client *Connection) ReadRegister() (*RegisterRequest, error) {
	msg, err := client.ReadMessage()
	if err != nil {
		return nil, err
	}
	var imsg interface{} = msg
	ack, ok := imsg.(*RegisterRequest)
	if !ok {
		return nil, fmt.Errorf("cannot cast message to register request")
	}
	return ack, nil
}

func (client *Connection) ReadTransferRequest() (*HttpTransferRequest, error) {
	msg, err := client.ReadMessage()
	if err != nil {
		return nil, err
	}
	var imsg interface{} = msg
	ack, ok := imsg.(*HttpTransferRequest)
	if !ok {
		return nil, fmt.Errorf("cannot cast message to transfer request")
	}
	return ack, nil
}

func (client *Connection) ReadTransferResponse() (*HttpTransferResponse, error) {
	msg, err := client.ReadMessage()
	if err != nil {
		return nil, err
	}
	var imsg interface{} = msg
	ack, ok := imsg.(*HttpTransferResponse)
	if !ok {
		return nil, fmt.Errorf("cannot cast message to transfer response")
	}
	return ack, nil
}


func (client *Connection) ReadHeartBeat() (*HeartBeat, error) {
	msg, err := client.ReadMessage()
	if err != nil {
		return nil, err
	}
	var imsg interface{} = msg
	ack, ok := imsg.(*HeartBeat)
	if !ok {
		return nil, fmt.Errorf("cannot cast message to transfer response")
	}
	return ack, nil
}


func (client *Connection) ReadMessage() (msg *Message, err error) {
	// 超时控制
	err = client.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return
	}
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

