package findSuccessor

import (
	nodeDefs "../utils/node_defs"
)

//FindSuccessor ... This function is used to find successor of node with ID targetID
func FindSuccessor(sponsoringNode *nodeDefs.Node, targetID int64) (successor *nodeDefs.Node) {
	if targetID > sponsoringNode.ChannelId && targetID <= sponsoringNode.Successor.ChannelId {
		successor = sponsoringNode.Successor
	} else {
		precedingNode := findClosestPrecedingNode(sponsoringNode, targetID)
		successor = FindSuccessor(precedingNode, targetID)
	}
	return successor
}

func findClosestPrecedingNode(sponsoringNode *nodeDefs.Node, targetID int64) (precedingNode *nodeDefs.Node) {

	for i := range sponsoringNode.FingerTable {
		if sponsoringNode.FingerTable[int64(i)].ChannelId > sponsoringNode.ChannelId && sponsoringNode.FingerTable[int64(i)].ChannelId < targetID {
			precedingNode = sponsoringNode.FingerTable[int64(i)]
		}
	}
	return
}
