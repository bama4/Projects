package findSuccessor

import (
	"math"

	nodeDefs "../utils/node_defs"
)

//FindSuccessor ... This function is used to find successor of node with ID targetID
func FindSuccessor(sponsoringNode *nodeDefs.Node, targetID int64, totalNodes int) (successor *nodeDefs.Node) {
	if targetID > sponsoringNode.ChannelId && targetID <= sponsoringNode.Successor.ChannelId {
		successor = sponsoringNode.Successor
	} else {
		precedingNode := findClosestPrecedingNode(sponsoringNode, targetID, totalNodes)
		successor = FindSuccessor(precedingNode, targetID, totalNodes)
	}
	return successor
}

func findClosestPrecedingNode(sponsoringNode *nodeDefs.Node, targetID int64, totalNodes int) (precedingNode *nodeDefs.Node) {

	for i := int(math.Log2(float64(totalNodes))) - 1; i >= 0; i-- {
		if sponsoringNode.FingerTable[int64(i)].ChannelId > sponsoringNode.ChannelId && sponsoringNode.FingerTable[int64(i)].ChannelId < targetID {
			precedingNode = sponsoringNode.FingerTable[int64(i)]
		}
	}
	return
}
