package leave_ring

import "fmt"
import "strconv"
import msg "./utils/message_defs"
import chord "./utils/node_defs"

func Leave_ring(sponsoring_node_id int64, node *chord.Node, mode String) {

	// Leaves orderly or immediate

	switch mode { 
	// Immediate just removes node, doesn't tell neighbors anything
	case "immediate":
		node.Predecessor = nil
		node.Successor = nil
	// Orderly requires dumping messages to successor before leaving
	case "orderly:
		// stuff to tell otehr nodes 

		network[sponsoring_node_id] <- "LEAVING"
		// stuff to dump data to other nodes
		// Loop through current nodes finger table
		// append it to successors
		//node.Successor.FingerTable = node.FingerTable
	
		// remove node from ring
		node.Predecessor = nil
		node.Successor = nil
		// dump fingr table to sponsoring node
		spnosoring_node.FingerTable = node.FingerTable
	default:
		// Immediate leave
		node.Predecessor = nil
		node.Successor = nil
	}

}
