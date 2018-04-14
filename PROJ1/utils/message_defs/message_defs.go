package message_defs

/*
This struct defines the json messages for the
CHORD protocol.

Fields:
do - The action for the node to take
sponsoring_node - The sponsoring node already in the coord ring for the action.
mode - The mode for leaving the CHORD ring
respond_to - The node (represented as a channel id) to direct the given action towards
data - The key/value pair in the ring representing a hash entry.
*/
type Message struct {
	Do             string `json:"do"`
	SponsoringNode string  `json:"sponsoring-node"`
	Mode           string `json:"mode"`
	RespondTo      string `json:"respond-to"`
	Data           Data   `json:"data"`
	TargetID       string  `json:"target-id"`
}

type Data struct {
	Key   string `json:"string"`
	Value string `json:"string"`
}
