package leave_ring

import "log"
import chord "../utils/node_defs"
import msg "../utils/message_defs"

// func SendDataToBucket(node_id int64, bucket_data int64){

/*
type Message struct {
        Do string `json:"do"`
        SponsoringNode int64 `json:"sponsoring-node"`
        Mode string `json:"mode"`
        RespondTo int64 `json:"respond-to"`
        TargetId int64 `json:target-id`
        Data Data `json:"data"`
}
*/

func Leave_ring(node *chord.Node, mode string) {

	// Leaves orderly or immediate

	switch mode { 
		case "immediate":
			node.Predecessor = -1
			node.Successor = -1
			node.FingerTable = nil
			log.Printf("\nNode: %d is leaving immediately\n", node.ChannelId)
			
		case "orderly":
			log.Printf("\nNode: %d is leaving orderly\n", node.ChannelId)
			// stuff to tell other nodes

//			network[node.Successor] <- "Node is leaving"
//			network[node.Predecessor] <- "Node is leaving"
			
			// Loop through current nodes DataTable to append to successor
			for k, v := range node.DataTable {

				var message = msg.Message {Do:"store-data", Mode:"orderly", TargetId:node.ChannelID,  Data:{k:v}}
			}
	
			// remove node from ring
			node.Predecessor = -1
			node.Successor = -1
			node.FingerTable = nil

		default:
			// Immediate leave
			node.Predecessor = -1
			node.Successor = -1
			node.FingerTable = nil
			log.Printf("\nNode: %d is leaving immediately\n", node.ChannelId)
	}

}
