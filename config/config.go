package config

import "flag"

var (
	TcpPort string
	HttpPort string
	Username string
	Password string
	HeartBeatRate int
	RegisterTimeout int
)


func init() {
	flag.StringVar(&TcpPort, "tcp", "8080", "tcp server port")
	flag.StringVar(&HttpPort, "http", "80", "http server port")
	flag.StringVar(&Username, "username", "admin", "username")
	flag.StringVar(&Password, "password", "bytedance", "password")
	flag.IntVar(&HeartBeatRate, "heart beat", 10, "heart beat")
	flag.IntVar(&RegisterTimeout, "register timeout", 30, "heart beat")
}