import "fmt"
import msg "./utils/message_defs"
import node "./utils/node_defs"
import "bytes"


//{"do": "put", "data": { "key" : "a key", "value" : "a value" }, "respond-to": "channel-id"}
// instructing the receipient node to store the given (key,value) pair in the appropriate ring node. 

// Node gets this message
// Use all the nodes fields

func find_biggest_node(finger_table *map[int64]*Node, node_id int64)(biggest int64) {

	var old_k := 0
	
	for k, v := range finger_table {
		if int64(k) < node_id {
			if int64(k) > old_k {
				biggest := k

			}
			old_k := k
		}	
	}
	return
}

func Put(data *msg.Data, respond_to int64) {

	// Biggest keyvalu
	var biggest = 0

	// Buffer

	var buffer bytes.Buffer

	buffer.WriteString("Putting data at node "

	// Get the node ID for data string
	var node_id = map_to_string_id(data.Key)

	// Check to see if this is the right node to store data
	if ChannelId == node_id {
		DataTable[data.Key] := data.Value
		return 
	}

	// Look in current node (this) fingertable	
	// FingerTable is map[int64]*node
	// Key is entry , value is node that has it
	for k, v := range FingerTable {
		if k == node_id {
			// We have a direct mapping for the key, go to this node
			v.DataTable[data.Key] := data.Value
			buffer.WriteString(string(node_id))
			network[respond_to] <- buffer.String()	
			return

		} else {
			// It's not in our finger table
			// Go to the biggest node without overshooting
			//var biggest = find_biggest_node(&FingerTable, node_id)		
			var closest_node = FindClosestPreceedingNode(&v, k)
			closest_node.DataTable[data.Key] := data.Value
			buffer.WriteString(string(closest_node.Key))
			network[respond_to] <- buffer.String()
		}		
	}

//	network[respond_to] <- "Putting data at node xx"

}
