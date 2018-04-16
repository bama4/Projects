package get

import "fmt"
import msg "./utils/message_defs"
import node "./utils/node_defs"




func Get(key String, respond-to int64)(value string) {

	var found = false
	var closest_node = nil
	// Get Node Data
	var node_id = map_to_string_id(key)

	// Check to see if we have it
	if ChannelID == node_id {
		value := DataTable[key]
	} else {
	
		// Check out Finger tables
		for k, v := range FingerTable {
			if k == node_id {
				value = DataTable[node_id]
				found = true
				break
			}
		}
	}
	
	// It's not in our finger table so we need to find the next highest node
	// without over shooting
	if found == false {
		closest_node = FindClosestPreceedingNode(&FingerTable, data.Key)
		value = closest_node.DataTable[data.Key]	
	}

	return
}
