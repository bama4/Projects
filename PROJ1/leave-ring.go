package leave-ring

import "fmt"
import "strconv"
import msg "./utils/message_defs"
import chord "./utils/chord_defs"

func leave-ring(sponsoring_node *ChordNode) {

	&sponsoring_node.predecessor = nil
	&sponsoring_node.successor = nil

}
