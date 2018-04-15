import "fmt"
import msg "./utils/message_defs"
import node "./utils/node_defs"


//{"do": "put", "data": { "key" : "a key", "value" : "a value" }, "respond-to": "channel-id"}
// instructing the receipient node to store the given (key,value) pair in the appropriate ring node. 
// Node gets this message

func Put(data *msg.Data, respond_to int64) {

	// Node gets data 
	//ring_nodes[respond_to].Data = data

	// map key from data.key to noe ID that is supposed to store
	// Find specific node in ring , if doesn't exist, map to successor

	// network[respond_to] <- "Put"


}
