package join-ring

import "fmt"
import "strconv"
import msg "./utils/message_defs"
import chord "./utils/node_defs"

func join-ring(node_id int, node *chord.Node){
	
	// Check global map to see if ring exists
	// if ring doesn't exist, create new ChordNode
	if ring_nodes == nil {
            node.Predecessor = nil
            node.Successor = nil
            ring_nodes[key] = node
	}
	// Add check to see if node id is in map/chord ring
	else {
	    node.Predecessor = nil
	    //node.Successor = find_successor()
	}

	ring_nodes[key] = node
}
