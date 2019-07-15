package tcp

import (
	"fmt"
	"proxy/config"
	"testing"
)

func TestClient(t *testing.T) {
	client, err := Connect("127.0.0.1", config.TcpPort, "/index",
		func(request *HttpTransferRequest) HttpTransferResponse {
		fmt.Printf("%#v", request)
		return HttpTransferResponse{
			Body: "hello world",
			Headers: map[string]string{
				"Content-Type":"application/json",
			},
		}
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(client.RemoteAddr().String())
	}
	select {

	}
}
