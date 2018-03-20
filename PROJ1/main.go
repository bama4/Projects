package main

import "fmt"
import msg "./utils/message_defs"
func main(){

	var response = &msg.Message {Do: "something", Data: msg.Data{Key:"sample key"}}
	fmt.Println(response.Data.Key)
}
