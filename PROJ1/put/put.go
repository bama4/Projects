import "fmt"
import msg "./utils/message_defs"
import node "./utils/node_defs"
import "strconv"


//{"do": "put", "data": { "key" : "a key", "value" : "a value" }, "respond-to": "channel-id"}
// instructing the receipient node to store the given (key,value) pair in the appropriate ring node. 

// Node gets this message
// Use all the nodes fields

func find_biggest(data *msg.Data, node_id int64)(biggest int64) {

	var bigger := 0
	var old_k := 0
	
	for k, v := range FingerTable {
		if int64(k) < node_id {
			bigger := k

		}	
	}
	
	return
}

func Put(data *msg.Data, respond_to int64) {

	// Biggest keyvalu
	var biggest = 0

	// Get the node ID for data string
	var node_id = map_to_string_id(data.Key)

	// Look in current node (this) fingertable	
	for k, v := range FingerTable {
		if k == node_id {
			// We have a direct mapping for the key, go to this node

		} else {

			// We don't have a direct mapping so we need to go to the next node in the finger table
			// We go to the node that is closest without over shooting
			if k < node_id {
				biggest = k


			}

			


		}		

	}

	// DataTable is also a map
//	ring_nodes[node_id].DataTable = data.Value
	// Need to do a lookup to a node and route to finger tables


	//ring_nodes[node_id].Value = data.Value
	
	network[respond_to] <- "Putting data at node xx"


}
