package chord_defs

import msg "../message_defs"

/*
This structure defines a chord node. 
Each node has a ChannelId (to simulate a network address)
and a Channel (to simulate a network socket) by
which the node recieves messages.

The Data entry contains the ambiguous data that may be associated
with a node as a key value pair.

The Finger Table represents the hash table consisting
of ChannelIds and the ChordNode that is on the path for that 
channel id.
*/
type ChordNode struct {
	ChannelId int64 `json:"channel_id"`
	Channel chan string `json:"data"`
	msg.Data `json:"data"`
	FingerTable map[int64]*ChordNode `json:"finger_table"`
}
