package main

import (
	_ "proxy/config"
	"proxy/http"
	"proxy/tcp"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 8)
	go http.Start()
	tcp.Start()
}
