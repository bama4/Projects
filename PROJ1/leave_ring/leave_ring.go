package leave_ring

import "log"
import chord "../utils/node_defs"
import msg "../utils/message_defs"
import "encoding/json"

func Leave_ring(node *chord.Node, mode string) {

	switch mode { 
		case "immediate":
			node.Predecessor = -1
			node.Successor = -1
			node.FingerTable = nil
			log.Printf("\nNode: %d is leaving immediately\n", node.ChannelId)
			
		case "orderly":
			log.Printf("\nNode: %d is leaving orderly\n", node.ChannelId)

			// Notify Successor and pRedecessor we are leaving
			var leaving_msg = "Leaving Ring"
			SendDataToNetwork(node.Successor, leaving_msg)
			SendDataToNetwork(node.Predecessor, leaving_msg)
			
			// Loop through current nodes DataTable to append to successor
			for k, v := range node.DataTable {
				var message = msg.Message {Do:"store-data-successor", Data:{k:v}}
				var string_msg, _ = json.Marshal(message)
		
				// Send Data to Successor
				SendDataToNetwork(node.ChannelID, string_message)
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
