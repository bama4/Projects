package join_ring

//import "fmt"
//import "strconv"
//import msg "../utils/message_defs"
import chord "../utils/node_defs"

// Gets sponsoring node ID to lookup
// Node object is the node that wants to join
func join_ring(sponsoring_node_id int, node *chord.Node){
	
    node.Predecessor = nil
    //node.Successor = find_successor(sponsoring_node_id)

}
