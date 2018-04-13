package leave_ring

import "fmt"
import "strconv"
import msg "./utils/message_defs"
import chord "./utils/node_defs"

func Leave_ring(sponsoring_node_id int64, node *chord.Node, mode String) {

	// Leaves orderly or immediate

	switch mode { 
		case "immediate":
			node.Predecessor = nil
			node.Successor = nil
		case "orderly:
			// stuff to tell otehr nodes 
			network[sponsoring_node_id] <- "LEAVING"
			
			// Loop through nodes fingertable to append to successor
			for k, v := range node.FingerTable {
				node.Successor.FingerTable[k] = v
			}
	
			// remove node from ring
			node.Predecessor = nil
			node.Successor = nil

		default:
			// Immediate leave
			node.Predecessor = nil
			node.Successor = nil
	}

}
