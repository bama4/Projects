package leave_ring

import "fmt"
import "strconv"
import msg "./utils/message_defs"
import chord "./utils/node_defs"

func leave-ring(sponsoring_node_id int, node *chord.Node, mode String) {

	// Leaves orderly or immediate

	switch mode { 
	// Immediate just removes node, doesn't tell neighbors anything
	case "immediate":
		node.Predecessor = nil
		node.Successor = nil
	// Orderly requires dumping messages to successor before leaving
	case "orderly:
		// stuff to tell otehr nodes 
		// stuff to dump data to other nodes
		node.Predecessor = nil
		node.Successor = nil
	default:
		fmt.Println("Error!")
	}

}
