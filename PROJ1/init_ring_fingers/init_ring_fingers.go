package init_ring_fingers

import chord "../utils/node_defs"
import "sync"

/*This is a lock that should eb used when writing to maps
*/
var map_lock = sync.Mutex{}

func Init_Ring_Fingers(node *chord.Node){

	if len(node.FingerTable) == 0	{
		//The finger tables are m bits long where m is the number of bits in the identifier.
		for i:=0; i < 64; i++ {
			//Should have 64 entries that increment in i + 2^i
			map_lock.Lock()
			node.FingerTable[int64(i)] = nil
			map_lock.Unlock()
		}
	}
}
