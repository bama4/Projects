import "fmt"
import msg "./utils/message_defs"
import node "./utils/node_defs"
import "bytes"


//{"do": "put", "data": { "key" : "a key", "value" : "a value" }, "respond-to": "channel-id"}
// instructing the receipient node to store the given (key,value) pair in the appropriate ring node. 

// Node gets this message
// Use all the nodes fields
/*
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
*/

func Put(data *msg.Data, respond_to int64) {


	var buffer bytes.Buffer
	var found = false
	var closest_node = nil

	buffer.WriteString("Putting data at node ")

	// Get the node ID for data string
	var node_id = map_to_string_id(data.Key)

	// Check to see if this is the right node to store data
	if ChannelId == node_id {
		DataTable[data.Key] := data.Value
		return 
	} else {
		for k, v := range FingerTable {
			if k == node_id {
				v.DataTable[data.Key] := data.Value
				buffer.WriteString(string(node_id))
				network[respond_to] <- buffer.String()
				found = true
				break
			}
		}	
		// It's not in our fingertable at all, so we need to go to the next highest node without
		// overshooting	
		if found == false {
			closest_node = FindClosestPreceedingNode(&FingerTable, data.Key)
			closest_node.DataTable[data.Key] := data.Value
			buffer.WriteString(string(closest_node.ChannelID))
			network[respond_to] <- buffer.String()
		}
	}
	//	network[respond_to] <- "Putting data at node xx"
}
