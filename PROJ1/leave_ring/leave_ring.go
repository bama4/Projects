package leave_ring

import "log"
import chord "../utils/node_defs"

func Leave_ring(node *chord.Node, mode string) {

	// Leaves orderly or immediate

	switch mode { 
		case "immediate":
			node.Predecessor = nil
			node.Successor = nil
			node.FingerTable = nil
			log.Printf("\nNode: %d is leaving immediately\n", node.ChannelId)
			
		case "orderly":
			log.Printf("\nNode: %d is leaving orderly\n", node.ChannelId)
			// stuff to tell other nodes
			
			
			// Loop through nodes fingertable to append to successor
			for k, v := range node.FingerTable {
				node.Successor.FingerTable[k] = v
			}
	
			// remove node from ring
			node.Predecessor = nil
			node.Successor = nil
			node.FingerTable = nil

		default:
			// Immediate leave
			node.Predecessor = nil
			node.Successor = nil
			node.FingerTable = nil
	}

}
