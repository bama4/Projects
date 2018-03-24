package main

import "fmt"
import "strconv"
import msg "./utils/message_defs"
import chord "./utils/chord_defs"

/*
This is the global "network" variable which is essentially a
map of all of the identifier/channel (aka network address/node) pairs
that we will have in our network.
*/
var network = make(map[int64](chan string))

func main(){

	var response = &msg.Message {Do: "something", Data: msg.Data{Key:"sample key"}}
	fmt.Println(response.Data.Key)

	var node = &chord.ChordNode {ChannelId: 121212}
	fmt.Println(strconv.FormatInt(node.ChannelId, 10))
	
}
