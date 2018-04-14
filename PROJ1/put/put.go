import "fmt"
import msg "./utils/message_defs"
import node "./utils/node_defs"


//{"do": "put", "data": { "key" : "a key", "value" : "a value" }, "respond-to": "channel-id"}
// instructing the receipient node to store the given (key,value) pair in the appropriate ring node. 
// Node gets this message

func Put(data *msg.Data, respond_to int64) {

	// Node gets data 
	node.Data = data


}
