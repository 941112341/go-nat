package tcp

import (
	"encoding/json"
	"fmt"
	"proxy/utils"
	"reflect"
	"testing"
)

func TestName(t *testing.T) {
	var i float32
	str := "6.12"
	_ = json.Unmarshal(utils.String2Splice(str), &i)
	fmt.Println(i)
}

func TestSet(t *testing.T) {
	var joker int
	Set(reflect.ValueOf(joker), "2")
	fmt.Print(joker)
}



func TestBinder(t *testing.T) {
	type User struct {
		Username string
		Password string `json:"password"`
		Id float64 `name:"id"`
	}
	str, _ := json.Marshal(User{Id:1, Password:"???"})
	request := HttpTransferRequest{
		Headers: map[string][]string{
			"Username": {"admin"},
		},
		FormQuery: map[string][]string{
			"id": { "1" },
		},
		Body: string(str),
	}

	var user User
	BinderReq(&request, &user)
	fmt.Printf("%#v", user)
}