package leave-ring

import "fmt"
import "strconv"
import msg "./utils/message_defs"
import chord "./utils/chord_defs"

func leave-ring(sponsoring_node *chord.ChordNode, mode String) {

	// Leaves orderly or immediate

	switch mode { 
	// Immediate just removes node, doesn't tell neighbors anything
	case "immediate":
		sponsoring_node.predecessor = nil
		sponsoring_node.successor = nil
	case "orderly:
		// stuff to tell otehr nodes 
		// stuff to dump data to other nodes
		sponsoring_node.predecessor = nil
		sponsoring_node.successor = nil
	default:
		fmt.Println("Error!")
	}

}
