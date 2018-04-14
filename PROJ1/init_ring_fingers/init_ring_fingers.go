package init_ring_fingers

import "log"
import chord "../utils/node_defs"
import "math"

func Init_Ring_Fingers(node *chord.Node){

	if len(node.FingerTable) == 0	{
		//Set first entry in the ring as the successor of the node
		log.Printf("\nJust initialized the finger table of Node %d\n", node.ChannelId)
		//The finger tables are m bits long where m is the number of bits in the identifier.
		for i:=0; i < 64; i++ {
			//Should have 64 entries that increment in i + 2^i
			node.FingerTable[int64(i) + int64(math.Pow(2.0, float64(i)))] = nil
		}
	}
}
