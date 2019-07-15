package tcp

import (
	"encoding/json"
	"log"
	"proxy/utils"
	"reflect"
	"strconv"
	"strings"
)


var typeMap = map[MessageType]reflect.Type{}

func init() {
	typeMap[ACK] = reflect.TypeOf(Ack{})
	typeMap[REGISTER] = reflect.TypeOf(RegisterRequest{})
	typeMap[TRANSFER_REQ] = reflect.TypeOf(HttpTransferRequest{})
	typeMap[TRANSFER_RESP] = reflect.TypeOf(HttpTransferResponse{})
	typeMap[HEART_BEAT] = reflect.TypeOf(HeartBeat{})
}


type HttpTransferRequest struct {
	Method string
	Headers map[string][]string
	FormQuery map[string][]string // form 和 query
	Body string
	ProxyMsg string
}

type HttpTransferResponse struct {
	Headers map[string]string
	Body string
}

type RegisterRequest struct {
	RegisterURL string
	Host string
}


type Ack struct {
	Ok bool
	Msg string
}

type HeartBeat struct {

}

type MessageType int

const (
	ACK MessageType = iota
	TRANSFER_REQ
	TRANSFER_RESP
	REGISTER
	HEART_BEAT
)

const Split =  "\r\n"

type Message struct {
	MType MessageType
	Content string
}

func (msg *Message) ToString() string {
	return strconv.FormatInt(int64(msg.MType), 10) + Split + msg.Content + Split
}

func (msg *Message) ProtoType() interface{} {
	i := reflect.New(typeMap[msg.MType]).Interface()
	err := json.Unmarshal(utils.String2Splice(msg.Content), i)
	if err != nil {
		log.Println(err)
	}
	return i
}

func NewMessage(msg string) *Message {
	arr := strings.Split(msg, Split)
	message := &Message{}
	mtype, err := strconv.Atoi(arr[0])
	if err != nil {
		log.Fatal("tcp:read message fail", err)
		return nil
	}
	message.MType = MessageType(mtype)
	message.Content = arr[1]
	return message
}

