package node_defs

import msg "../message_defs"

/*
This structure defines a node. 
Each goroutine that is listening on a channel will use this structure to store its
node information.

The Data entry contains the ambiguous data that may be associated
with a node as a key value pair.

The Finger Table represents the hash table consisting
of ChannelIds and the ChordNode that is on the path for that 
channel id.

The is_in_ring field indicates if this node is part of the chord ring.
*/
type Node struct {
	ChannelId int64 `json:"channel_id"`
	DataTable map[string]string `json:"data"`
	FingerTable map[int64]*Node `json:"finger_table"`
	Successor *Node `json:"successor"`
	Predecessor *Node `json:"predecessor"`
}
