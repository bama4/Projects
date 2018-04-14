package init_ring_fingers

import "log"
import chord "../utils/node_defs"


func Init_Ring_Fingers(node *chord.Node){

	if len(node.FingerTable) == 0	{
		//Set first entry in the ring as the successor of the node
		log.Printf("\nJust initialized the finger table of Node %d\n", node.ChannelId)
		node.FingerTable[node.Successor.ChannelId] = node.Successor
	}
}
