package join-ring

import "fmt"
import "strconv"
import msg "./utils/message_defs"
import chord "./utils/node_defs"

func join-ring(sponsoring_node *chord.Node){
	
	// Check global map to see if ring exists
	// if ring doesn't exist, create new ChordNode
	if len(chord.ring_nodes) == 0 {
            sponsoring_node.Predecessor = nil
            sponsoring_node.Successor = nil
            chord.ring_nodes[key] = sponsoring_node
	}
	else {
	    sponsoring_node.Predecessor = nil
	    sponsoring_node.Successor = find_successor()
	}

	chord.ring_nodes[key] = sponsoring_node
}
