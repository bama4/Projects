package join-ring

import "fmt"
import "strconv"
import msg "./utils/message_defs"
import chord "./utils/chord_defs"

func join-ring(sponsoring_node *ChordNode){
	
	// Check global map to see if ring exists
	// if ring doesn't exist, create new ChordNode
	if len(ring_nodes) == 0 {
            &sponsoring_node.predecessor = nil
            &sponsoring_node.successor = nil
            ring_nodes[key] = sponsoring_node
	}
	else {
	    &sponsoring_node.predecessor = nil
	    &sponsoring_node.successor = find_successor()
	}

	ring_nodes[key] = sponsoring_node
}
