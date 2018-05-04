package get

import msg "./utils/message_defs"
import node "./utils/node_defs"

func Get(data *msg.Data, respond-to int64, node_obj *node.Node)(value string) {

	// Get Node Data
	var node_id = map_to_string_id(key)
	
	// It's not in our finger table so we need to find the next highest node
	var closest_node = FindClosestPreceedingNode(&node_obj.FingerTable, data.Key)
	value = closest_node.DataTable[data.Key]	

	return
}
