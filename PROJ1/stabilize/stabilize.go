package stabilize

import (
	nodeDefs "../utils/node_defs"
)

//Stabilize ...
func Stabilize(node *nodeDefs.Node) {
	x := node.Successor.Predecessor
	if x.ChannelId > node.ChannelId && x.ChannelId < node.Successor.ChannelId {
		node.Successor = x
	}
	//Notify(node, successor)
}
