import "fmt"
import msg "./utils/message_defs"
import node "./utils/node_defs"




func Get(key String, respond-to int64)(value string) {

	// Get Node Data
	var node_id = map_to_string_id(key)

	// Check to see if we have it
	if ChannelID == node_id {

		value := DataTable[key]

	}

	// Lookup key in own finger table
	for k, v := range FingerTable {
		if k == node_id {
			value := DataTable[key]
			return
		} else {

			var closest_node = find_biggest_node(FingerTable, node_id)
			value := closest_node.DataTable[key]
			return	
		}
	}

	return
}
