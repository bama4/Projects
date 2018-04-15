package init_ring_fingers

import chord "../utils/node_defs"
import "sync"

/*This is a lock that should eb used when writing to maps
*/
var map_lock = sync.Mutex{}

func Init_Ring_FingerTable(node *chord.Node, number_of_network_nodes int){

	if len(node.FingerTable) == 0	{
		//The finger tables are m bits long where m is the number of bits in the identifier.
		for i:=0; i < number_of_network_nodes; i++ {
			//Should have N entries for a ring or up to 2^N nodes
			map_lock.Lock()
			node.FingerTable[int64(i)] = nil
			map_lock.Unlock()
		}
	}
}
