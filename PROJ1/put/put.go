package put

import msg "./utils/message_defs"
import node "./utils/node_defs"
//import "bytes"


//{"do": "put", "data": { "key" : "a key", "value" : "a value" }, "respond-to": "channel-id"}
// instructing the receipient node to store the given (key,value) pair in the appropriate ring node. 

// func FindClosestPreceedingNode(node_obj *node.Node, target_id int64)

func Put(data *msg.Data, respond_to int64, node_obj *node.Node) {

	// Get the node ID for data string
	var node_id = map_to_string_id(data.Key)

	// Check to see if this is the right node to store data
	if ChannelId == node_id {
		DataTable[data.Key] := data.Value
	} else {
		// FindPReceedingNode
		var closest_node = FindClosestPreceedingNode(&node_obj.FingerTable, data.Key)

		// Put data in node
		closest_node.DataTable[data.Key] := data.Value
	}
}
