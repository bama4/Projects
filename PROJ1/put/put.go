import "fmt"
import msg "./utils/message_defs"
import node "./utils/node_defs"


//{"do": "put", "data": { "key" : "a key", "value" : "a value" }, "respond-to": "channel-id"}
// instructing the receipient node to store the given (key,value) pair in the appropriate ring node. 
// Node gets this message

func Put(data *msg.Data, respond_to int64) {

	// Get the node ID for data string
	var node_id = map_to_string_id(data.Key)

	// Node gets data 
	// DataTable is also a map
//	ring_nodes[node_id].DataTable = data.Value
	// Need to do a lookup to a node and route to finger tables

	for i := 0; i < len(ring_nodes); i++ {
		if ring_nodes[i] != node_id {
			// Look in finger table
			for j :=0; j < len(ring_nodes[i].FingerTable); j++ {
				


			}	

		}	

	}


	//ring_nodes[node_id].Value = data.Value
	
	// network[respond_to] <- "Put"


}
