package leave-ring

import "fmt"
import "strconv"
import msg "./utils/message_defs"
import chord "./utils/node_defs"

func leave-ring(node_id int, sponsoring_node *chord.Node, mode String) {

	// Leaves orderly or immediate

	switch mode { 
	// Immediate just removes node, doesn't tell neighbors anything
	case "immediate":
		sponsoring_node.Predecessor = nil
		sponsoring_node.Successor = nil
	// Orderly requires dumping messages to successor before leaving
	case "orderly:
		// stuff to tell otehr nodes 
		// stuff to dump data to other nodes
		sponsoring_node.Predecessor = nil
		sponsoring_node.Successor = nil
	default:
		fmt.Println("Error!")
	}

}
